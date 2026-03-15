package ai

import "testing"

func TestFindModelInfoWithNilLoginMetadataDoesNotPanic(t *testing.T) {
	client := &AIClient{}

	if got := client.findModelInfo(""); got != nil {
		t.Fatalf("expected nil model info for empty model id, got %#v", got)
	}
}
