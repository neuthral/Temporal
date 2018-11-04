package storj_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/storj"
)

const (
	defaultAccessKey = "insecure-dev-access-key"
	defaultSecretKey = "insecure-dev-secret-key"
	defaultEndpoint  = "127.0.0.1:7777"
)

func TestStorj(t *testing.T) {
	opts := &storj.Opts{
		AccessKey: defaultAccessKey,
		SecretKey: defaultSecretKey,
		Endpoint:  defaultEndpoint,
	}
	client, err := storj.NewStorjClient(opts)
	if err != nil {
		t.Fatal(err)
	}
	if err = client.Mini.MakeBucket(
		map[string]string{"name": "testbucket", "location": "testLocation"},
	); err != nil {
		t.Fatal(err)
	}
}
