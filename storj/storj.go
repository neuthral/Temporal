package storj

import (
	"context"
	"errors"
	"io"

	"github.com/RTradeLtd/Temporal/mini"
	"github.com/storj/storj/pkg/overlay"
	"google.golang.org/grpc"
	"storj.io/storj/pkg/pb"
)

// Client is our interface with storj
type Client struct {
	Mini *mini.MinioManager
	OC   pb.OverlayClient
}

// Opts are our configuration options for storj
type Opts struct {
	AccessKey  string
	SecretKey  string
	Endpoint   string
	ServerAddr string
}

// NewStorjClient is used to initialize our storj client
func NewStorjClient(opts *Opts) (*Client, error) {
	mm, err := mini.NewMinioManager(
		opts.Endpoint, opts.AccessKey, opts.SecretKey, false,
	)
	if err != nil {
		return nil, err
	}
	sClient, err := overlay.NewClient(opts.ServerAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &Client{Mini: mm, OC: sClient}, nil
}

// Store is used to store a given object in storj
func (c *Client) Store(name string, reader io.Reader, objectSize int64, opts mini.PutObjectOptions) (int64, error) {
	return c.Mini.PutObject(name, reader, objectSize, opts)
}

// Lookup finds a nodes address from the network
func (c *Client) Lookup(ctx context.Context, req *pb.LookupRequest, opts ...grpc.CallOption) (*pb.LookupResponse, error) {
	return nil, errors.New("not yet implemented")
	//return c.OC.Lookup(ctx, req, opts...)
}
