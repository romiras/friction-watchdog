package mcp

import (
	"bufio"
	"context"
	"encoding/json"
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

		// 1. Обработка Handshake (Обязательно для MCP)
		if req.Method == "initialize" {
			log.Println("[MCP] Получен запрос initialize. Отправляем capabilities...")
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

			// Отвечаем на обязательный пинг initialized (подтверждение от клиента)
		} else if req.Method == "notifications/initialized" {
			log.Println("[MCP] Сессия установлена.")

			// 2. Листинг ресурсов
		} else if req.Method == "resources/list" {
			log.Println("[MCP] Агент запрашивает список ресурсов...")
			resp := ResponseRPC{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"resources": []map[string]interface{}{
						{
							"uri":         "productivity://friction-report",
							"name":        "Friction Report",
							"mimeType":    "application/json",
							"description": "Отчет о трении в процессе разработки (простой, ошибки)",
						},
					},
				},
			}
			encoder.Encode(resp)

			// 3. Обработка чтения ресурса
		} else if req.Method == "resources/read" && req.Params.URI == "productivity://friction-report" {
			log.Println("[MCP] Агент запрашивает статус. Отдаем Friction Report...")
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

			// 4. Подписки на ресурсы (фиктивная обработка для совместимости)
		} else if req.Method == "resources/subscribe" || req.Method == "resources/unsubscribe" {
			log.Printf("[MCP] Получен запрос %s для %s", req.Method, req.Params.URI)
			resp := ResponseRPC{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  map[string]interface{}{}, // Успешный ответ
			}
			encoder.Encode(resp)

			// 5. Tools & Prompts
		} else if req.Method == "tools/list" {
			log.Println("[MCP] Агент запрашивает список инструментов...")
			resp := ResponseRPC{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"tools": []map[string]interface{}{
						{
							"name":        "reset_timer",
							"description": "Сбрасывает таймер неактивности и счетчик ошибок.",
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
				log.Println("[MCP] Выполняется инструмент reset_timer...")
				stateManager.Reset()
				resp := ResponseRPC{
					JSONRPC: "2.0",
					ID:      req.ID,
					Result: map[string]interface{}{
						"content": []map[string]interface{}{
							{
								"type": "text",
								"text": "Таймер успешно сброшен.",
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
			log.Println("[MCP] Агент запрашивает список промптов...")
			resp := ResponseRPC{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"prompts": []map[string]interface{}{
						{
							"name":        "coaching_mode",
							"description": "Инструкции для включения режима Coaching Mode.",
						},
					},
				},
			}
			encoder.Encode(resp)

		} else if req.Method == "prompts/get" {
			if req.Params.Name == "coaching_mode" {
				log.Println("[MCP] Отдаем промпт coaching_mode...")
				resp := ResponseRPC{
					JSONRPC: "2.0",
					ID:      req.ID,
					Result: map[string]interface{}{
						"description": "Инструкции для включения режима Coaching Mode.",
						"messages": []map[string]interface{}{
							{
								"role": "user",
								"content": map[string]interface{}{
									"type": "text",
									"text": "Ты теперь действуешь как коуч по продуктивности. Проанализируй Friction Report и, если есть проблемы, предложи помощь. Будь краток и конструктивен.",
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

			// 6. Ping (проверка активности)
		} else if req.Method == "ping" {
			resp := ResponseRPC{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  map[string]interface{}{},
			}
			encoder.Encode(resp)

			// 6. Обработка неизвестных методов (чтобы не вешать клиента таймаутом)
		} else if req.ID != nil {
			log.Printf("[MCP] Метод не поддерживается: %s", req.Method)
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
