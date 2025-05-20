package gossip

import (
	"context"
	"log"

	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// Gossip wraps a simple libp2p Gossipsub topic.
type Gossip struct {
	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription
}

func New(ctx context.Context) (*Gossip, error) {
	h, err := libp2p.New()
	if err != nil {
		return nil, err
	}
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, err
	}
	topic, err := ps.Join("nebula-edge-func")
	if err != nil {
		return nil, err
	}
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}
	g := &Gossip{
		ctx:   ctx,
		ps:    ps,
		topic: topic,
		sub:   sub,
	}
	go g.listen()
	return g, nil
}

func (g *Gossip) listen() {
	for {
		msg, err := g.sub.Next(g.ctx)
		if err != nil {
			return
		}
		log.Printf("gossip recv %d bytes", len(msg.Data))
		// Future: integrate with agent to store function code.
	}
}

// Broadcast publishes data on the topic.
func (g *Gossip) Broadcast(data []byte) error {
	return g.topic.Publish(g.ctx, data)
}
