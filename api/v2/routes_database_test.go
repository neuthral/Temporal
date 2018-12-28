package v2

import (
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
)

func Test_API_Routes_Database(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// setup fake mock clients
	fakeLens := &mocks.FakeIndexerAPIClient{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}

	api, _, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
	if err != nil {
		t.Fatal(err)
	}

	// test database specific uploads
	// /api/v2/database/uploads/testuser
	var interfaceAPIResp interfaceAPIResponse
	if err := sendRequest(
		api, "GET", "/api/v2/database/uploads", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad api status code from api/v2/database/uploads")
	}

	// test get encrypted uploads - no network
	// /api/v2/frontend/uploads/encrypted
	if err := sendRequest(
		api, "GET", "/api/v2/database/uploads/encrypted", 200, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test get encrypted uploads - with network
	// /api/v2/frontend/uploads/encrypted
	if err := sendRequest(
		api, "GET", "/api/v2/database/uploads/encrypted/testNetwork", 200, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}
}
