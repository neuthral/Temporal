package relay

import (
	"context"
	"time"

	inet "gx/ipfs/QmPtFaR7BWHLAjSwLh9kXcyrgTzDpuhcWLkx8ioa9RMYnx/go-libp2p-net"
	ma "gx/ipfs/QmRKLtwMw131aK7ugC3G7ybpumMz78YrJe5dzneyindvG1/go-multiaddr"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
)

var _ inet.Notifiee = (*RelayNotifiee)(nil)

type RelayNotifiee Relay

func (r *Relay) Notifiee() inet.Notifiee {
	return (*RelayNotifiee)(r)
}

func (n *RelayNotifiee) Relay() *Relay {
	return (*Relay)(n)
}

func (n *RelayNotifiee) Listen(net inet.Network, a ma.Multiaddr)      {}
func (n *RelayNotifiee) ListenClose(net inet.Network, a ma.Multiaddr) {}
func (n *RelayNotifiee) OpenedStream(net inet.Network, s inet.Stream) {}
func (n *RelayNotifiee) ClosedStream(net inet.Network, s inet.Stream) {}

func (n *RelayNotifiee) Connected(s inet.Network, c inet.Conn) {
	if n.Relay().Matches(c.RemoteMultiaddr()) {
		return
	}

	go func(id peer.ID) {
		ctx, cancel := context.WithTimeout(n.ctx, time.Second)
		defer cancel()

		canhop, err := n.Relay().CanHop(ctx, id)
		if err != nil {
			log.Debugf("Error testing relay hop: %s", err.Error())
			return
		}

		if canhop {
			log.Debugf("Discovered hop relay %s", id.Pretty())
			n.mx.Lock()
			n.relays[id] = struct{}{}
			n.mx.Unlock()
			n.host.ConnManager().TagPeer(id, "relay-hop", 2)
		}
	}(c.RemotePeer())
}

func (n *RelayNotifiee) Disconnected(s inet.Network, c inet.Conn) {}
