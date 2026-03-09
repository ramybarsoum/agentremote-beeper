package openclaw

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"maunium.net/go/mautrix/bridgev2"
)

func TestOpenClawLoginStartUsesAuthModeSelect(t *testing.T) {
	login := &OpenClawLogin{
		User:      &bridgev2.User{},
		Connector: &OpenClawConnector{br: &bridgev2.Bridge{}},
	}

	step, err := login.Start(context.Background())
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if step.StepID != openClawLoginStepAuthMode {
		t.Fatalf("unexpected first step id: %q", step.StepID)
	}
	if step.UserInputParams == nil || len(step.UserInputParams.Fields) != 1 {
		t.Fatalf("expected a single select field, got %#v", step.UserInputParams)
	}
	field := step.UserInputParams.Fields[0]
	if field.Type != bridgev2.LoginInputFieldTypeSelect {
		t.Fatalf("expected select field, got %q", field.Type)
	}
	if len(field.Options) != 3 {
		t.Fatalf("expected three auth mode options, got %#v", field.Options)
	}
}

func TestOpenClawLoginSubmitUserInputReturnsModeSpecificFields(t *testing.T) {
	cases := []struct {
		name         string
		inputMode    string
		wantStepID   string
		wantFieldIDs []string
	}{
		{
			name:         "no auth",
			inputMode:    "No auth",
			wantStepID:   openClawLoginStepCredentialsNoAuth,
			wantFieldIDs: []string{"url", "label"},
		},
		{
			name:         "token",
			inputMode:    "Token",
			wantStepID:   openClawLoginStepCredentialsToken,
			wantFieldIDs: []string{"url", "token", "label"},
		},
		{
			name:         "password",
			inputMode:    "Password",
			wantStepID:   openClawLoginStepCredentialsPass,
			wantFieldIDs: []string{"url", "password", "label"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			login := &OpenClawLogin{
				User:      &bridgev2.User{},
				Connector: &OpenClawConnector{br: &bridgev2.Bridge{}},
			}
			if _, err := login.Start(context.Background()); err != nil {
				t.Fatalf("Start returned error: %v", err)
			}

			step, err := login.SubmitUserInput(context.Background(), map[string]string{"auth_mode": tc.inputMode})
			if err != nil {
				t.Fatalf("SubmitUserInput returned error: %v", err)
			}
			if step.StepID != tc.wantStepID {
				t.Fatalf("unexpected step id: got %q want %q", step.StepID, tc.wantStepID)
			}
			if step.UserInputParams == nil {
				t.Fatalf("expected user input params for %s", tc.name)
			}

			gotFieldIDs := make([]string, 0, len(step.UserInputParams.Fields))
			for _, field := range step.UserInputParams.Fields {
				gotFieldIDs = append(gotFieldIDs, field.ID)
			}
			if len(gotFieldIDs) != len(tc.wantFieldIDs) {
				t.Fatalf("unexpected field count: got %#v want %#v", gotFieldIDs, tc.wantFieldIDs)
			}
			for i := range gotFieldIDs {
				if gotFieldIDs[i] != tc.wantFieldIDs[i] {
					t.Fatalf("unexpected field ids: got %#v want %#v", gotFieldIDs, tc.wantFieldIDs)
				}
			}
		})
	}
}

func TestNormalizeOpenClawAuthCredentials(t *testing.T) {
	if _, _, err := normalizeOpenClawAuthCredentials("token", map[string]string{}); err == nil {
		t.Fatal("expected token mode without token to fail")
	}
	if _, _, err := normalizeOpenClawAuthCredentials("password", map[string]string{}); err == nil {
		t.Fatal("expected password mode without password to fail")
	}
	token, password, err := normalizeOpenClawAuthCredentials("none", map[string]string{
		"token":    "abc",
		"password": "secret",
	})
	if err != nil {
		t.Fatalf("none auth mode returned error: %v", err)
	}
	if token != "" || password != "" {
		t.Fatalf("expected none auth mode to clear credentials, got token=%q password=%q", token, password)
	}
}

func TestOpenClawLoginSubmitUserInputPairingRequiredReturnsWaitStep(t *testing.T) {
	login := &OpenClawLogin{
		User:      &bridgev2.User{},
		Connector: &OpenClawConnector{br: &bridgev2.Bridge{}},
		preflight: func(context.Context, string, string, string) (string, error) {
			return "", &gatewayRPCError{
				Method:     "connect",
				Message:    "pairing required",
				DetailCode: "PAIRING_REQUIRED",
				RequestID:  "req-123",
			}
		},
	}
	if _, err := login.Start(context.Background()); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if _, err := login.SubmitUserInput(context.Background(), map[string]string{"auth_mode": "Token"}); err != nil {
		t.Fatalf("auth mode SubmitUserInput returned error: %v", err)
	}

	step, err := login.SubmitUserInput(context.Background(), map[string]string{
		"url":   "ws://127.0.0.1:18789",
		"token": "shared-token",
	})
	if err != nil {
		t.Fatalf("credentials SubmitUserInput returned error: %v", err)
	}
	if step.Type != bridgev2.LoginStepTypeDisplayAndWait {
		t.Fatalf("unexpected step type: %q", step.Type)
	}
	if step.StepID != openClawLoginStepPairingWait {
		t.Fatalf("unexpected step id: %q", step.StepID)
	}
	if step.DisplayAndWaitParams == nil || step.DisplayAndWaitParams.Type != bridgev2.LoginDisplayTypeNothing {
		t.Fatalf("unexpected display-and-wait params: %#v", step.DisplayAndWaitParams)
	}
	if !strings.Contains(step.Instructions, "req-123") {
		t.Fatalf("expected request ID in instructions, got %q", step.Instructions)
	}
	if login.step != openClawLoginStatePairingWait {
		t.Fatalf("unexpected login state: %q", login.step)
	}
	if login.pending == nil || login.pending.requestID != "req-123" {
		t.Fatalf("unexpected pending login: %#v", login.pending)
	}
}

func TestOpenClawLoginWaitReturnsStillWaitingStepOnContextDone(t *testing.T) {
	login := &OpenClawLogin{
		User:      &bridgev2.User{},
		Connector: &OpenClawConnector{br: &bridgev2.Bridge{}},
		step:      openClawLoginStatePairingWait,
		pending: &openClawPendingLogin{
			gatewayURL: "ws://127.0.0.1:18789",
			authMode:   "token",
			token:      "shared-token",
			requestID:  "req-456",
		},
		waitUntil: time.Now().Add(time.Minute),
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	step, err := login.Wait(ctx)
	if err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}
	if step.StepID != openClawLoginStepPairingWait {
		t.Fatalf("unexpected step id: %q", step.StepID)
	}
	if !strings.Contains(step.Instructions, "Still waiting") {
		t.Fatalf("expected still waiting instructions, got %q", step.Instructions)
	}
}

func TestOpenClawLoginWaitMapsNonPairingErrors(t *testing.T) {
	login := &OpenClawLogin{
		User:       &bridgev2.User{},
		Connector:  &OpenClawConnector{br: &bridgev2.Bridge{}},
		step:       openClawLoginStatePairingWait,
		pollEvery:  time.Millisecond,
		returnWait: time.Second,
		waitFor:    time.Second,
		pending: &openClawPendingLogin{
			gatewayURL: "ws://127.0.0.1:18789",
			authMode:   "token",
			token:      "shared-token",
			requestID:  "req-789",
		},
		preflight: func(context.Context, string, string, string) (string, error) {
			return "", &gatewayRPCError{
				Method:     "connect",
				Message:    "token mismatch",
				DetailCode: "AUTH_TOKEN_MISMATCH",
			}
		},
	}

	_, err := login.Wait(context.Background())
	if err == nil {
		t.Fatal("expected Wait to return an error")
	}
	var respErr bridgev2.RespError
	if !errors.As(err, &respErr) {
		t.Fatalf("expected RespError, got %T", err)
	}
	if respErr.StatusCode != 403 {
		t.Fatalf("unexpected status code: %d", respErr.StatusCode)
	}
}
