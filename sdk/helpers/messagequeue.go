package helpers

import (
	"sync"
)

// MessageQueue serializes message processing per room, ensuring only one
// handler runs at a time for each room ID.
type MessageQueue struct {
	mu     sync.Mutex
	active map[string]chan struct{}
}

// NewMessageQueue creates a new MessageQueue.
func NewMessageQueue() *MessageQueue {
	return &MessageQueue{
		active: make(map[string]chan struct{}),
	}
}

// Enqueue runs handler for the given room, waiting for any in-progress handler
// to finish first. Multiple Enqueue calls for the same room are serialized.
func (q *MessageQueue) Enqueue(roomID string, handler func()) {
	q.waitForRoom(roomID)
	q.acquireRoom(roomID)
	defer q.ReleaseRoom(roomID)
	handler()
}

// AcquireRoom marks a room as active. Returns true if the room was not already
// active, false if it was (caller should wait or skip).
func (q *MessageQueue) AcquireRoom(roomID string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, ok := q.active[roomID]; ok {
		return false
	}
	q.active[roomID] = make(chan struct{})
	return true
}

// ReleaseRoom marks a room as no longer active.
func (q *MessageQueue) ReleaseRoom(roomID string) {
	q.mu.Lock()
	ch, ok := q.active[roomID]
	if ok {
		delete(q.active, roomID)
	}
	q.mu.Unlock()
	if ok && ch != nil {
		close(ch)
	}
}

// HasActiveRoom returns true if the given room is currently being processed.
func (q *MessageQueue) HasActiveRoom(roomID string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	_, ok := q.active[roomID]
	return ok
}

func (q *MessageQueue) waitForRoom(roomID string) {
	for {
		q.mu.Lock()
		ch, ok := q.active[roomID]
		q.mu.Unlock()
		if !ok {
			return
		}
		<-ch
	}
}

func (q *MessageQueue) acquireRoom(roomID string) {
	q.mu.Lock()
	q.active[roomID] = make(chan struct{})
	q.mu.Unlock()
}
