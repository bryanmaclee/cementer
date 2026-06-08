// Package hub fans out already-committed readings to all connected WebSocket
// clients. It is intentionally ignorant of the transport: it deals only in
// Subscribers that hold a buffered Send channel. The WebSocket layer creates a
// Subscriber, registers it, and pumps its Send channel to the socket.
//
// The reliability rule lives here: a slow subscriber whose buffer fills is DROPPED
// rather than allowed to back up ingestion. Data is already durable in SQLite, so
// a dropped client simply reconnects and backfills.
package hub

import "context"

// Subscriber is one connected client from the hub's point of view.
type Subscriber struct {
	// Send carries serialized messages to the client's write pump. The hub closes
	// it when the subscriber is unregistered or dropped.
	Send chan []byte
}

// NewSubscriber makes a subscriber with a buffered send channel.
func NewSubscriber(buffer int) *Subscriber {
	if buffer <= 0 {
		buffer = 64
	}
	return &Subscriber{Send: make(chan []byte, buffer)}
}

type Hub struct {
	register   chan *Subscriber
	unregister chan *Subscriber
	broadcast  chan []byte
	subs       map[*Subscriber]struct{}
}

func New() *Hub {
	return &Hub{
		register:   make(chan *Subscriber),
		unregister: make(chan *Subscriber, 16),
		broadcast:  make(chan []byte, 256),
		subs:       make(map[*Subscriber]struct{}),
	}
}

// Run processes register/unregister/broadcast until ctx is cancelled. Launch it in
// its own goroutine.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for s := range h.subs {
				delete(h.subs, s)
				close(s.Send)
			}
			return
		case s := <-h.register:
			h.subs[s] = struct{}{}
		case s := <-h.unregister:
			if _, ok := h.subs[s]; ok {
				delete(h.subs, s)
				close(s.Send)
			}
		case msg := <-h.broadcast:
			for s := range h.subs {
				select {
				case s.Send <- msg:
				default:
					// Slow client: drop it. Ingestion is never blocked by a client.
					delete(h.subs, s)
					close(s.Send)
				}
			}
		}
	}
}

// Register adds a subscriber.
func (h *Hub) Register(s *Subscriber) { h.register <- s }

// Unregister removes a subscriber (idempotent from the caller's side).
func (h *Hub) Unregister(s *Subscriber) {
	select {
	case h.unregister <- s:
	default:
		// unregister buffer full; the Run loop will reap on the next broadcast drop.
	}
}

// Broadcast enqueues a message for all subscribers. It never blocks the caller
// (the store writer): if the hub is overloaded the live message is dropped, since
// the data is already durably stored and clients can backfill.
func (h *Hub) Broadcast(msg []byte) {
	select {
	case h.broadcast <- msg:
	default:
	}
}
