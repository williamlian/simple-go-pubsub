package pubsub

import (
	"log"

	"github.com/google/uuid"
)

// PubSub is the basic interface for in-memory pub sub
type PubSub interface {
	Publish(topic string, data string)
	Subscribe(topic string, sub Subscriber) uuid.UUID
}

// Subscriber is the inteface subscriber must implement
type Subscriber interface {
	OnData(data string) (unsubscribe bool)
}

type subscriberType struct {
	active chan bool
	send   chan string
}

type pubSubImpl map[uuid.UUID]subscriberType

type pubSub struct {
	topics map[string]pubSubImpl
}

// NewPubSub creates a new pub sub component
func NewPubSub() PubSub {
	return &pubSub{
		topics: map[string]pubSubImpl{},
	}
}

/* Potential Memory Leak: this pub sub require publish to detect closed consumer.
   Let's say at item 5 the consumer closes,
   There is already a item 6 blocked on the sending channel
   So block 6 will be consumed by the empty loop after the onData() == false.
   And the consumer will remain blocked by the data channel waiting the publisher to close.
   Only after item 7, the publisher will close the data channel and clean up the map.

   So, in summary, all subscriber will exist till N+2 publish events to their topic.

   Do not use this in a dynamic topic environment where topic changes quickly and each of them
   will only have couple of publish. Because this will leave a lot of zombie goroutines in
   the memory!
*/
func (p *pubSub) Publish(topic string, data string) {
	subscribers, ok := p.topics[topic]
	if !ok || len(subscribers) == 0 {
		log.Printf("topic %s does not have any subscribers", topic)
		return
	}
	for k, s := range subscribers {
		// first make sure subscriber is still alive
		select {
		case _, ok = <-s.active:
			if !ok {
				// we have something from active which means the channel is closed
				log.Printf("[%s] subscriber closed, remove.\n", k)
				close(s.send)
				delete(subscribers, k)
			} else {
				panic("unexpected receive from active channel")
			}
			continue
		default:
			// subscriber is active
			s.send <- data
			log.Printf("[%s] data sent.\n", k)
		}
	}
	return
}

func (p *pubSub) Subscribe(topic string, subscriber Subscriber) uuid.UUID {
	if _, ok := p.topics[topic]; !ok {
		p.topics[topic] = pubSubImpl{}
	}
	id := uuid.New()
	dataChannel := make(chan string)
	activeChannel := make(chan bool)
	defer func() {
		if r := recover(); r != nil {
			/* this method should never panic, but just in case
			we should close the active channel properly to prevent
			blocking */
			close(activeChannel)
		}
	}()
	p.topics[topic][id] = subscriberType{
		active: activeChannel,
		send:   dataChannel,
	}
	go _onData(id, activeChannel, dataChannel, subscriber.OnData)
	return id
}

// goroutine
// if onData crashed, it will auto unsubscribe
func _onData(id uuid.UUID, activeChannel chan bool, dataChannel chan string, onData func(string) bool) {
	defer func() {
		log.Printf("Subscriber %s exit", id)
		if r := recover(); r != nil {
			close(activeChannel)
		}
	}()
	for data := range dataChannel {
		shouldUnsub := onData(data)
		if shouldUnsub {
			close(activeChannel)
			/* Important: the consumer should stay alive for a while to clear
			the sending channel until the producer close it to unblock the producer */
			for range dataChannel {
			}
			return
		}
	}
}
