package mcp

type BaseMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
}

type NotificationRPC struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		URI string `json:"uri"`
	} `json:"params"`
}

type RequestRPC struct {
	BaseMessage
	Params struct {
		URI       string                 `json:"uri,omitempty"`
		Name      string                 `json:"name,omitempty"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
	} `json:"params,omitempty"`
}

type ResponseRPC struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *ErrorRPC   `json:"error,omitempty"`
}

type ErrorRPC struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
