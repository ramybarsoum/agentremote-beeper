package openclaw

import (
	"errors"
	"strings"
	"testing"
	"time"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/status"
)

func TestMapOpenClawLoginErrorPairingRequired(t *testing.T) {
	err := mapOpenClawLoginError(&gatewayRPCError{
		Method:     "connect",
		Message:    "pairing required",
		DetailCode: "PAIRING_REQUIRED",
		RequestID:  "req-123",
	})
	var respErr bridgev2.RespError
	if !errors.As(err, &respErr) {
		t.Fatalf("expected RespError, got %T", err)
	}
	if respErr.StatusCode != 403 {
		t.Fatalf("unexpected status code: %d", respErr.StatusCode)
	}
	if got := respErr.Error(); got == "" || !containsAll(got, []string{"pairing", "req-123", "openclaw devices approve req-123"}) {
		t.Fatalf("unexpected error text: %q", got)
	}
}

func TestMapOpenClawLoginErrorAuthFailure(t *testing.T) {
	err := mapOpenClawLoginError(&gatewayRPCError{
		Method:     "connect",
		Message:    "token mismatch",
		DetailCode: "AUTH_TOKEN_MISMATCH",
	})
	var respErr bridgev2.RespError
	if !errors.As(err, &respErr) {
		t.Fatalf("expected RespError, got %T", err)
	}
	if respErr.StatusCode != 403 {
		t.Fatalf("unexpected status code: %d", respErr.StatusCode)
	}
}

func TestClassifyOpenClawConnectionErrorPairingRequired(t *testing.T) {
	state, retry := classifyOpenClawConnectionError(&gatewayRPCError{
		Method:     "connect",
		Message:    "pairing required",
		DetailCode: "PAIRING_REQUIRED",
		RequestID:  "req-123",
	}, time.Second)
	if retry {
		t.Fatal("expected pairing-required error to stop retries")
	}
	if state.StateEvent != status.StateBadCredentials {
		t.Fatalf("unexpected state event: %s", state.StateEvent)
	}
	if state.Error != openClawPairingRequiredError {
		t.Fatalf("unexpected state error: %s", state.Error)
	}
	if got := state.Info["request_id"]; got != "req-123" {
		t.Fatalf("unexpected request id info: %#v", state.Info)
	}
}

func TestClassifyOpenClawConnectionErrorAuthFailure(t *testing.T) {
	state, retry := classifyOpenClawConnectionError(&gatewayRPCError{
		Method:     "connect",
		Message:    "token mismatch",
		DetailCode: "AUTH_TOKEN_MISMATCH",
	}, time.Second)
	if retry {
		t.Fatal("expected auth failure to stop retries")
	}
	if state.StateEvent != status.StateBadCredentials {
		t.Fatalf("unexpected state event: %s", state.StateEvent)
	}
	if state.Error != openClawAuthFailedError {
		t.Fatalf("unexpected state error: %s", state.Error)
	}
}

func containsAll(value string, subs []string) bool {
	for _, sub := range subs {
		if !strings.Contains(value, sub) {
			return false
		}
	}
	return true
}
