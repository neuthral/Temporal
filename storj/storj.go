package storj

import "github.com/RTradeLtd/Temporal/mini"

// Client is our interface with storj
type Client struct {
	Mini *mini.MinioManager
}

// Opts are our configuration options for storj
type Opts struct {
	AccessKey string
	SecretKey string
	Endpoint  string
}

// NewStorjClient is used to initialize our storj client
func NewStorjClient(opts *Opts) (*Client, error) {
	mm, err := mini.NewMinioManager(
		opts.Endpoint, opts.AccessKey, opts.SecretKey, false,
	)
	if err != nil {
		return nil, err
	}
	return &Client{Mini: mm}, nil
}
