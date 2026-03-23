package rpc

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
)

func TestMapHandlerError_nil(t *testing.T) {
	if MapHandlerError(nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestMapHandlerError_passThroughConnectError(t *testing.T) {
	orig := connect.NewError(connect.CodeInvalidArgument, errors.New("bad input"))
	out := MapHandlerError(orig)
	if out != orig {
		t.Fatal("expected same connect error instance")
	}
}

func TestMapHandlerError_wrapsPlainError(t *testing.T) {
	plain := errors.New("sql: connection refused secret details")
	out := MapHandlerError(plain)
	if out == nil {
		t.Fatal("expected error")
	}
	ce, ok := out.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", out)
	}
	if ce.Code() != connect.CodeInternal {
		t.Fatalf("code: got %v want %v", ce.Code(), connect.CodeInternal)
	}
	if ce.Message() == plain.Error() {
		t.Fatal("must not expose underlying error message")
	}
}

func TestMapHandlerError_vertexPublisherModelNoAccess(t *testing.T) {
	plain := errors.New(`Publisher Model foo does not have access`)
	out := MapHandlerError(plain)
	if out == nil {
		t.Fatal("expected error")
	}
	ce, ok := out.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", out)
	}
	if ce.Code() != connect.CodeFailedPrecondition {
		t.Fatalf("code: got %v want %v", ce.Code(), connect.CodeFailedPrecondition)
	}
}
