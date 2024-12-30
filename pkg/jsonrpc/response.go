package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// ResponseMsg is a struct allowing to encode/decode a JSON-RPC response body
type ResponseMsg struct {
	Version string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

func Unmarshal(msg *ResponseMsg, res interface{}) error {
	if msg.Error == nil && msg.Result == nil {
		return fmt.Errorf("invalid JSON-RPC response missing both result and error")
	}

	if msg.Error != nil {
		errMsg := new(ErrorMsg)
		err := json.Unmarshal(msg.Error, errMsg)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON-RPC error message %v", string(msg.Error))
		}
		return errMsg
	}

	if msg.Result != nil && res != nil {
		err := json.Unmarshal(msg.Result, res)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON-RPC result %v into %T (%v)", string(msg.Result), res, err)
		}
		return nil
	}

	return nil
}
