package connector

import "testing"

func TestIntegrationPortalRoomType(t *testing.T) {
	t.Run("visible rooms stay ai", func(t *testing.T) {
		if got := integrationPortalRoomType(nil); got != "ai" {
			t.Fatalf("expected ai room type, got %q", got)
		}
	})

	t.Run("internal module rooms are ai prefixed", func(t *testing.T) {
		meta := &PortalMetadata{
			ModuleMeta: map[string]any{
				"cron": map[string]any{"is_internal_room": true},
			},
		}
		if got := integrationPortalRoomType(meta); got != "ai-cron" {
			t.Fatalf("expected ai-cron room type, got %q", got)
		}
	})

	t.Run("heartbeat rooms are ai prefixed", func(t *testing.T) {
		meta := &PortalMetadata{
			ModuleMeta: map[string]any{
				"heartbeat": map[string]any{"is_internal_room": true},
			},
		}
		if got := integrationPortalRoomType(meta); got != "ai-heartbeat" {
			t.Fatalf("expected ai-heartbeat room type, got %q", got)
		}
	})
}
