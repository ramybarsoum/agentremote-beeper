package openclaw

import (
	"errors"
	"strings"
	"testing"

	"maunium.net/go/mautrix/bridgev2"
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

func containsAll(value string, subs []string) bool {
	for _, sub := range subs {
		if !strings.Contains(value, sub) {
			return false
		}
	}
	return true
}
