package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"friction-watchdog/internal/state"
)

func StartHandler(ctx context.Context, stateManager *state.Manager) {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	lines := make(chan string)
	go func() {
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()

	for {
		var line string
		select {
		case <-ctx.Done():
			log.Println("[MCP] Handler shutting down...")
			return
		case l, ok := <-lines:
			if !ok {
				log.Println("[MCP] Stdin closed, exiting.")
				return
			}
			line = l
		}

		var req RequestRPC
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			log.Printf("Parse error: %v", err)
			continue
		}

		// 1. Handshake handling (Required for MCP)
		if req.Method == "initialize" {
			log.Println("[MCP] Received initialize request. Sending capabilities...")
			resp := ResponseRPC{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"protocolVersion": "2024-11-05",
					"capabilities": map[string]interface{}{
						"resources": map[string]interface{}{
							"subscribe": true,
						},
						"tools":   map[string]interface{}{},
						"prompts": map[string]interface{}{},
					},
					"serverInfo": map[string]string{
						"name":    "friction-watchdog",
						"version": "1.0.0",
					},
				},
			}
			encoder.Encode(resp)

			// Respond to mandatory initialized notification (client confirmation)
		} else if req.Method == "notifications/initialized" {
			log.Println("[MCP] Session established.")

			// 2. Resource listing
		} else if req.Method == "resources/list" {
			log.Println("[MCP] Agent requests resource list...")
			resp := ResponseRPC{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"resources": []map[string]interface{}{
						{
							"uri":         "productivity://friction-report",
							"name":        "Friction Report",
							"mimeType":    "application/json",
							"description": "Productivity friction report (idle, errors)",
						},
					},
				},
			}
			encoder.Encode(resp)

			// 3. Resource reading
		} else if req.Method == "resources/read" && req.Params.URI == "productivity://friction-report" {
			log.Println("[MCP] Agent requests status. Sending Friction Report...")
			resp := ResponseRPC{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"contents": []map[string]interface{}{
						{
							"uri":      "productivity://friction-report",
							"mimeType": "application/json",
							"text":     stateManager.GetReport(),
						},
					},
				},
			}
			encoder.Encode(resp)

			// 4. Resource subscriptions (mock handling for compatibility)
		} else if req.Method == "resources/subscribe" || req.Method == "resources/unsubscribe" {
			log.Printf("[MCP] Received request %s for %s", req.Method, req.Params.URI)
			resp := ResponseRPC{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  map[string]interface{}{}, // Successful response
			}
			encoder.Encode(resp)

			// 5. Tools & Prompts
		} else if req.Method == "tools/list" {
			log.Println("[MCP] Agent requests tool list...")
			resp := ResponseRPC{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"tools": []map[string]interface{}{
						{
							"name":        "reset_timer",
							"description": "Resets the inactivity timer and error count.",
							"inputSchema": map[string]interface{}{
								"type":       "object",
								"properties": map[string]interface{}{},
							},
						},
					},
				},
			}
			encoder.Encode(resp)

		} else if req.Method == "tools/call" {
			if req.Params.Name == "reset_timer" {
				log.Println("[MCP] Executing reset_timer tool...")
				stateManager.Reset()
				resp := ResponseRPC{
					JSONRPC: "2.0",
					ID:      req.ID,
					Result: map[string]interface{}{
						"content": []map[string]interface{}{
							{
								"type": "text",
								"text": "Timer successfully reset.",
							},
						},
					},
				}
				encoder.Encode(resp)
			} else {
				resp := ResponseRPC{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error: &ErrorRPC{
						Code:    -32601,
						Message: "Tool not found: " + req.Params.Name,
					},
				}
				encoder.Encode(resp)
			}

		} else if req.Method == "prompts/list" {
			log.Println("[MCP] Agent requests prompt list...")
			resp := ResponseRPC{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"prompts": []map[string]interface{}{
						{
							"name":        "coaching_mode",
							"description": "Instructions for enabling Coaching Mode.",
						},
					},
				},
			}
			encoder.Encode(resp)

		} else if req.Method == "prompts/get" {
			if req.Params.Name == "coaching_mode" {
				log.Println("[MCP] Sending coaching_mode prompt...")
				report := stateManager.GetReport()
				promptText := fmt.Sprintf(
					"You are a productivity coach. "+
						"The developer triggered a friction alert. Diagnostic data: %s. "+
						"Based on this data: if trigger_reason is 'error_loop', ask one focused "+
						"question about the failing command. If 'idle_timeout', ask if they are "+
						"blocked or need to decompose the current task. Be brief and constructive.",
					report,
				)
				resp := ResponseRPC{
					JSONRPC: "2.0",
					ID:      req.ID,
					Result: map[string]interface{}{
						"description": "Instructions for enabling Coaching Mode.",
						"messages": []map[string]interface{}{
							{
								"role": "system",
								"content": map[string]interface{}{
									"type": "text",
									"text": promptText,
								},
							},
						},
					},
				}
				encoder.Encode(resp)
			} else {
				resp := ResponseRPC{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error: &ErrorRPC{
						Code:    -32601,
						Message: "Prompt not found: " + req.Params.Name,
					},
				}
				encoder.Encode(resp)
			}

			// 6. Ping (activity check)
		} else if req.Method == "ping" {
			resp := ResponseRPC{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  map[string]interface{}{},
			}
			encoder.Encode(resp)

			// 6. Unknown method handling (to avoid hanging the client)
		} else if req.ID != nil {
			log.Printf("[MCP] Method not supported: %s", req.Method)
			resp := ResponseRPC{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &ErrorRPC{
					Code:    -32601, // Method not found
					Message: "Method not found: " + req.Method,
				},
			}
			encoder.Encode(resp)
		}
	}
}
