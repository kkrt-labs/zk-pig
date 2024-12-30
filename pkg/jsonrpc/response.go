package jsonrpc

import (
	"encoding/json"
	"fmt"
	"io"
)

// ResponseMsg is a struct allowing to encode/decode a JSON-RPC response body
type ResponseMsg struct {
	Version string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

// Unmarshal unmarshals a JSON-RPC response result into a given interface
// If the response contains an error, it will be unmarshaled into an ErrorMsg and returned
func (msg *ResponseMsg) Unmarshal(res interface{}) error {
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

// DecodeResponseMsg decodes a JSON-RPC response message from an io.Reader
func DecodeResponseMsg(r io.Reader) (*ResponseMsg, error) {
	msg := new(ResponseMsg)
	err := json.NewDecoder(r).Decode(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode into JSON-RPC response message: %v", err)
	}
	return msg, nil
}
