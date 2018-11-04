package storj_test

import (
	"os"
	"testing"

	"github.com/RTradeLtd/Temporal/mini"

	"github.com/RTradeLtd/Temporal/storj"
)

const (
	defaultAccessKey  = "insecure-dev-access-key"
	defaultSecretKey  = "insecure-dev-secret-key"
	defaultEndpoint   = "127.0.0.1:7777"
	defaultServerAddr = "127.0.0.1:7780"
	defaultNodeID     = "3hr1PINJUEG5qM5SiScsQKqcBCxYlEORRFPKQq9p8AA"
	defaultFile       = "../test/config.json"
	defaultObjectName = "randomobjectname"
)

func TestStorj(t *testing.T) {
	//t.Skip()
	opts := &storj.Opts{
		AccessKey:  defaultAccessKey,
		SecretKey:  defaultSecretKey,
		Endpoint:   defaultEndpoint,
		ServerAddr: defaultServerAddr,
	}
	client, err := storj.NewStorjClient(opts)
	if err != nil {
		t.Fatal(err)
	}
	if err = client.Mini.MakeBucket(
		map[string]string{"name": "testbucket10", "location": "testLocation"},
	); err != nil {
		t.Error(err)
	}
	file, err := os.Open(defaultFile)
	if err != nil {
		t.Fatal(err)
	}
	fileStats, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = client.Store(
		defaultObjectName,
		file,
		fileStats.Size(),
		mini.PutObjectOptions{
			Bucket: "testbucket10",
		},
	); err != nil {
		t.Fatal(err)
	}
}
