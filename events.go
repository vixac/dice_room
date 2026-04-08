package main

import "sync"

// Broadcaster manages SSE subscriber channels per room.
type Broadcaster struct {
	mu          sync.Mutex
	subscribers map[string][]chan string
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{subscribers: make(map[string][]chan string)}
}

// Subscribe registers a new channel for the given room and returns it.
func (b *Broadcaster) Subscribe(roomID string) chan string {
	ch := make(chan string, 16)
	b.mu.Lock()
	b.subscribers[roomID] = append(b.subscribers[roomID], ch)
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes the channel from the room and closes it.
func (b *Broadcaster) Unsubscribe(roomID string, ch chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subscribers[roomID]
	for i := range subs {
		if subs[i] == ch {
			b.subscribers[roomID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	close(ch)
}

// Send delivers a message to all subscribers of a room, skipping any that are full.
func (b *Broadcaster) Send(roomID string, msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subscribers[roomID] {
		select {
		case ch <- msg:
		default:
		}
	}
}
