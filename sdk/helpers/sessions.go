package helpers

import (
	"sync"

	"maunium.net/go/mautrix/bridgev2"
)

// SessionTracker tracks the mapping between sessions and portals.
// This is useful for bridges that need to know which portal a session belongs to.
type SessionTracker struct {
	mu               sync.RWMutex
	sessionToPortal  map[string]*bridgev2.Portal
	portalToSessions map[string]map[string]struct{}
}

// NewSessionTracker creates a new SessionTracker.
func NewSessionTracker() *SessionTracker {
	return &SessionTracker{
		sessionToPortal:  make(map[string]*bridgev2.Portal),
		portalToSessions: make(map[string]map[string]struct{}),
	}
}

// Register associates a session ID with a portal.
func (t *SessionTracker) Register(sessionID string, portal *bridgev2.Portal) {
	if sessionID == "" || portal == nil {
		return
	}
	portalID := string(portal.ID)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sessionToPortal[sessionID] = portal
	sessions, ok := t.portalToSessions[portalID]
	if !ok {
		sessions = make(map[string]struct{})
		t.portalToSessions[portalID] = sessions
	}
	sessions[sessionID] = struct{}{}
}

// Unregister removes a session ID from tracking.
func (t *SessionTracker) Unregister(sessionID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	portal, ok := t.sessionToPortal[sessionID]
	if !ok {
		return
	}
	delete(t.sessionToPortal, sessionID)
	if portal != nil {
		portalID := string(portal.ID)
		if sessions, exists := t.portalToSessions[portalID]; exists {
			delete(sessions, sessionID)
			if len(sessions) == 0 {
				delete(t.portalToSessions, portalID)
			}
		}
	}
}

// GetPortal returns the portal associated with a session ID, or nil.
func (t *SessionTracker) GetPortal(sessionID string) *bridgev2.Portal {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.sessionToPortal[sessionID]
}

// GetSessions returns all session IDs associated with a portal ID.
func (t *SessionTracker) GetSessions(portalID string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	sessions := t.portalToSessions[portalID]
	if len(sessions) == 0 {
		return nil
	}
	result := make([]string, 0, len(sessions))
	for s := range sessions {
		result = append(result, s)
	}
	return result
}

// HasSessions returns true if the given portal has any active sessions.
func (t *SessionTracker) HasSessions(portalID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.portalToSessions[portalID]) > 0
}

// Clear removes all tracked sessions.
func (t *SessionTracker) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sessionToPortal = make(map[string]*bridgev2.Portal)
	t.portalToSessions = make(map[string]map[string]struct{})
}
