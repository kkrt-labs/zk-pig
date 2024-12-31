package jsonrpcmock

import (
	"fmt"

	jsonrpc "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc"
	"go.uber.org/mock/gomock"
)

type matcher struct {
	match func(req *jsonrpc.Request) bool
	msg   string
}

func (m *matcher) Matches(x interface{}) bool {
	req, ok := x.(*jsonrpc.Request)
	if !ok {
		return false
	}

	return m.match(req)
}

func (m *matcher) String() string {
	return m.msg
}

func HasVersion(v string) gomock.Matcher {
	return &matcher{
		match: func(req *jsonrpc.Request) bool { return req.Version == v },
		msg:   fmt.Sprintf("Request should have version %q", v),
	}
}

func HasID(id interface{}) gomock.Matcher {
	return &matcher{
		match: func(req *jsonrpc.Request) bool { return req.ID == id },
		msg:   fmt.Sprintf("Request should have version %v", id),
	}
}
