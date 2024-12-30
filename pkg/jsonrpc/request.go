package jsonrpc

import (
	"encoding/json"
)

var null = json.RawMessage("null")

// Request allows to manipulate a JSON-RPC request
type Request struct {
	Version string
	Method  string
	ID      interface{}
	Params  interface{}
}

// RequestMsg is a struct allowing to encode/decode a JSON-RPC request body
type RequestMsg struct {
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

// MarshalJSON
func (msg *Request) MarshalJSON() ([]byte, error) {
	raw := new(RequestMsg)

	raw.Version = msg.Version
	raw.Method = msg.Method
	raw.ID = msg.ID

	if msg.Params != nil {
		b, err := json.Marshal(msg.Params)
		if err != nil {
			return nil, err
		}
		raw.Params = b
	} else {
		copy(raw.Params, null)
	}

	return json.Marshal(raw)
}
