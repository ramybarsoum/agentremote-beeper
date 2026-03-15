package ai

import (
	"testing"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/bridgev2/networkid"

	"github.com/beeper/agentremote"
)

func testUserLoginWithMeta(loginID networkid.UserLoginID, meta *UserLoginMetadata) *bridgev2.UserLogin {
	return &bridgev2.UserLogin{
		UserLogin: &database.UserLogin{
			ID:       loginID,
			Metadata: meta,
		},
	}
}

func TestAIClientNeedsRebuild(t *testing.T) {
	existing := &AIClient{
		apiKey:    "secret",
		UserLogin: testUserLoginWithMeta("existing", &UserLoginMetadata{Provider: " OpenAI ", BaseURL: "https://api.example.com/v1/"}),
	}

	if aiClientNeedsRebuild(existing, "secret", &UserLoginMetadata{Provider: "openai", BaseURL: "https://api.example.com/v1"}) {
		t.Fatal("expected no rebuild when key/provider/base URL are equivalent")
	}
	if !aiClientNeedsRebuild(existing, "other-key", &UserLoginMetadata{Provider: "openai", BaseURL: "https://api.example.com/v1"}) {
		t.Fatal("expected rebuild when API key changes")
	}
	if !aiClientNeedsRebuild(existing, "secret", &UserLoginMetadata{Provider: "openrouter", BaseURL: "https://api.example.com/v1"}) {
		t.Fatal("expected rebuild when provider changes")
	}
	if !aiClientNeedsRebuild(existing, "secret", &UserLoginMetadata{Provider: "openai", BaseURL: "https://api.other.example.com/v1"}) {
		t.Fatal("expected rebuild when base URL changes")
	}
	if !aiClientNeedsRebuild(nil, "secret", &UserLoginMetadata{Provider: "openai"}) {
		t.Fatal("expected rebuild when no existing client is cached")
	}
}

func TestLoadAIUserLoginMissingAPIKeyEvictsCacheAndSetsBrokenClient(t *testing.T) {
	loginID := networkid.UserLoginID("login-1")
	oc := &OpenAIConnector{
		clients: map[networkid.UserLoginID]bridgev2.NetworkAPI{},
	}
	cachedLogin := testUserLoginWithMeta(loginID, nil)
	oc.clients[loginID] = newBrokenLoginClient(cachedLogin, "cached")

	login := testUserLoginWithMeta(loginID, nil)
	if err := oc.loadAIUserLogin(login, &UserLoginMetadata{Provider: ProviderOpenAI}); err != nil {
		t.Fatalf("loadAIUserLogin returned error: %v", err)
	}
	if _, ok := oc.clients[loginID]; ok {
		t.Fatal("expected cached client to be evicted when API key is missing")
	}
	if login.Client == nil {
		t.Fatal("expected broken login client")
	}
	if _, ok := login.Client.(*agentremote.BrokenLoginClient); !ok {
		t.Fatalf("expected broken login client type, got %T", login.Client)
	}
}
