package bitswap

import (
	"context"
	"sync"
	"testing"

	bsmsg "gx/ipfs/QmUYXFM46WgGs5AScfL4FSZXa9p5nAhddueyM5auAVZGCQ/go-bitswap/message"
	bsnet "gx/ipfs/QmUYXFM46WgGs5AScfL4FSZXa9p5nAhddueyM5auAVZGCQ/go-bitswap/network"

	testutil "gx/ipfs/QmPuhRE325DR8ChNcFtgd6F1eANCHy1oohXZPpYop4xsK6/go-testutil"
	delay "gx/ipfs/QmUe1WCHkQaz4UeNKiHDUBV2T6i9prc3DniqyHPXyfGaUq/go-ipfs-delay"
	blocks "gx/ipfs/QmWoXtvgC8inqFkAATB7cp2Dax7XBi9VDvSg9RCCZufmRk/go-block-format"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	mockrouting "gx/ipfs/QmdmWkx54g7VfVyxeG8ic84uf4G6Eq1GohuyKA3XDuJ8oC/go-ipfs-routing/mock"
)

func TestSendMessageAsyncButWaitForResponse(t *testing.T) {
	net := VirtualNetwork(mockrouting.NewServer(), delay.Fixed(0))
	responderPeer := testutil.RandIdentityOrFatal(t)
	waiter := net.Adapter(testutil.RandIdentityOrFatal(t))
	responder := net.Adapter(responderPeer)

	var wg sync.WaitGroup

	wg.Add(1)

	expectedStr := "received async"

	responder.SetDelegate(lambda(func(
		ctx context.Context,
		fromWaiter peer.ID,
		msgFromWaiter bsmsg.BitSwapMessage) {

		msgToWaiter := bsmsg.New(true)
		msgToWaiter.AddBlock(blocks.NewBlock([]byte(expectedStr)))
		waiter.SendMessage(ctx, fromWaiter, msgToWaiter)
	}))

	waiter.SetDelegate(lambda(func(
		ctx context.Context,
		fromResponder peer.ID,
		msgFromResponder bsmsg.BitSwapMessage) {

		// TODO assert that this came from the correct peer and that the message contents are as expected
		ok := false
		for _, b := range msgFromResponder.Blocks() {
			if string(b.RawData()) == expectedStr {
				wg.Done()
				ok = true
			}
		}

		if !ok {
			t.Fatal("Message not received from the responder")
		}
	}))

	messageSentAsync := bsmsg.New(true)
	messageSentAsync.AddBlock(blocks.NewBlock([]byte("data")))
	errSending := waiter.SendMessage(
		context.Background(), responderPeer.ID(), messageSentAsync)
	if errSending != nil {
		t.Fatal(errSending)
	}

	wg.Wait() // until waiter delegate function is executed
}

type receiverFunc func(ctx context.Context, p peer.ID,
	incoming bsmsg.BitSwapMessage)

// lambda returns a Receiver instance given a receiver function
func lambda(f receiverFunc) bsnet.Receiver {
	return &lambdaImpl{
		f: f,
	}
}

type lambdaImpl struct {
	f func(ctx context.Context, p peer.ID, incoming bsmsg.BitSwapMessage)
}

func (lam *lambdaImpl) ReceiveMessage(ctx context.Context,
	p peer.ID, incoming bsmsg.BitSwapMessage) {
	lam.f(ctx, p, incoming)
}

func (lam *lambdaImpl) ReceiveError(err error) {
	// TODO log error
}

func (lam *lambdaImpl) PeerConnected(p peer.ID) {
	// TODO
}
func (lam *lambdaImpl) PeerDisconnected(peer.ID) {
	// TODO
}
