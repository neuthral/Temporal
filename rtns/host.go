package rtns

import (
	"context"
	"encoding/base64"
	"time"

	path "gx/ipfs/QmZErC2Ay6WuGi96CPg316PwitdwgLo6RxZRqVjJjRj2MR/go-path"

	ci "gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	config "gx/ipfs/QmYyzmMnhNTtoXx5ttgUaRdHHckYnQWjPL98hgLAR2QLDD/go-ipfs-config"
	ds "gx/ipfs/Qmf4xQhNomPNhrtZc67qSnfJSjxjXs9LWvknJtSXwimPrM/go-datastore"

	"github.com/ipfs/go-ipfs/core"
	repo "github.com/ipfs/go-ipfs/repo"
)

// Publisher provides a helper to publish IPNS records
type Publisher struct {
	host *core.IpfsNode
}

// Opts is used to configure our connection
type Opts struct {
	PK ci.PrivKey
}

// NewPublisher is used to generate our IPNS publisher
func NewPublisher(pk ci.PrivKey, permanent bool, swarmAddrs ...string) (*Publisher, error) {
	pid, err := peer.IDFromPrivateKey(pk)
	if err != nil {
		return nil, err
	}
	pkBytes, err := pk.Bytes()
	if err != nil {
		return nil, err
	}
	// generate a blank config
	c := config.Config{}
	// popular config with necessary defaults
	c.Bootstrap = config.DefaultBootstrapAddresses
	c.Addresses.Swarm = swarmAddrs
	c.Identity.PeerID = pid.Pretty()
	c.Identity.PrivKey = base64.StdEncoding.EncodeToString(pkBytes)
	// generate a null datastore, as we just want to publish records
	d := ds.NewNullDatastore()
	// create a mock repo to feed into our node
	repoMock := repo.Mock{
		C: c,
		D: d,
	}
	// create a new node
	host, err := core.NewNode(context.Background(), &core.BuildCfg{
		Online:    true,
		Permanent: permanent,
		Repo:      &repoMock,
	})
	if err != nil {
		return nil, err
	}
	return &Publisher{
		host: host,
	}, nil
}

// PublishWithEOL is used to publish an IPNS record with non default lifetime values
func (p *Publisher) PublishWithEOL(ctx context.Context, pk ci.PrivKey, content string, eol time.Time) error {
	return p.host.Namesys.PublishWithEOL(ctx, pk, path.FromString(content), eol)
}
