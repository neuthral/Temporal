// Code generated by ifacemaker. DO NOT EDIT.

package rtfs

import (
	"context"
	"io"
	"time"

	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
)

// Manager provides functions for interacting with IPFS
type Manager interface {
	// NodeAddress returns the node the manager is connected to
	NodeAddress() string
	// Add is a wrapper used to add a file to IPFS
	// currently until https://github.com/ipfs/go-ipfs/issues/5376 it is added with no pin
	// thus a manual pin must be triggered afterwards
	Add(r io.Reader) (string, error)
	// DagPut is used to store data as an ipld object
	DagPut(data interface{}, encoding, kind string) (string, error)
	// DagGet is used to get an ipld object
	DagGet(cid string, out interface{}) error
	// Cat is used to get cat an ipfs object
	Cat(cid string) ([]byte, error)
	// Stat is used to retrieve the stats about an object
	Stat(hash string) (*ipfsapi.ObjectStats, error)
	// Pin is a wrapper method to pin a hash to the local node,
	// but also alert the rest of the local nodes to pin
	// after which the pin will be sent to the cluster
	Pin(hash string) error
	// CheckPin checks whether or not a pin is present
	CheckPin(hash string) (bool, error)
	// Publish is used for fine grained control over IPNS record publishing
	Publish(contentHash, keyName string, lifetime, ttl time.Duration, resolve bool) (*ipfsapi.PublishResponse, error)
	// PubSubPublish is used to publish a a message to the given topic
	PubSubPublish(topic string, data string) error
	// CustomRequest is used to make a custom request
	CustomRequest(ctx context.Context, url, commad string, opts map[string]string, args ...string) (*ipfsapi.Response, error)
}
