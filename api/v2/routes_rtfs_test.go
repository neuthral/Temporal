package v2

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
)

func Test_API_Routes_IPFS_Public(t *testing.T) {
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

	api, testRecorder, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	// add a file normally
	// /api/v2/ipfs/public/file/add
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("file", "../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	fh, err := os.Open("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	if _, err = io.Copy(fileWriter, fh); err != nil {
		t.Fatal(err)
	}
	bodyWriter.Close()
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v2/ipfs/public/file/add", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues := url.Values{}
	urlValues.Add("hold_time", "5")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /api/v2/ipfs/public/file/add")
	}
	var apiResp apiResponse
	// unmarshal the response
	bodyBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/ipfs/public/file/add")
	}
	hash = apiResp.Response

	// add a file advanced
	// /api/v2/ipfs/public/file/add/advanced
	bodyBuf = &bytes.Buffer{}
	bodyWriter = multipart.NewWriter(bodyBuf)
	fileWriter, err = bodyWriter.CreateFormFile("file", "../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	fh, err = os.Open("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	if _, err = io.Copy(fileWriter, fh); err != nil {
		t.Fatal(err)
	}
	bodyWriter.Close()
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/ipfs/public/file/add/advanced", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	urlValues.Add("passphrase", "password123")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /api/v2/ipfs/public/file/add/advanced")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/ipfs/public/file/add/advanced")
	}

	// test pinning - missing hold_time
	// /api/v2/ipfs/public/pin
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/public/pin/"+hash, 400, nil, nil, &apiResp,
	); err != nil {
		t.Fatal(err)
	}

	// test pinning
	// /api/v2/ipfs/public/pin
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/public/pin/"+hash, 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from  /api/v2/ipfs/public/pin")
	}

	// test pubsub publish - missing message
	// /api/v2/ipfs/pubsub/publish/topic
	var mapAPIResp mapAPIResponse
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/public/pubsub/publish/foo", 400, nil, nil, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}

	// test pubsub publish
	// /api/v2/ipfs/pubsub/publish/topic
	urlValues = url.Values{}
	urlValues.Add("message", "bar")
	mapAPIResp = mapAPIResponse{}
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/public/pubsub/publish/foo", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api status code from  /api/v2/pubsub/publish/topic")
	}
	if mapAPIResp.Response["topic"] != "foo" {
		t.Fatal("bad response")
	}
	if mapAPIResp.Response["message"] != "bar" {
		t.Fatal("bad response")
	}

	// test object stat
	// /api/v2/ipfs/stat
	var interfaceAPIResp interfaceAPIResponse
	if err := sendRequest(
		api, "GET", "/api/v2/ipfs/public/stat/"+hash, 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/public/stat")
	}

	// test get dag
	// /api/v2/ipfs/public/dag
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/api/v2/ipfs/public/dag/"+hash, 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/public/dag/")
	}

	// test download
	// /api/v2/ipfs/utils/download
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/utils/download/"+hash, 200, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test public network beam - missing source_network
	// /api/v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "public")
	urlValues.Add("content_hash", hash)
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/utils/laser/beam", 400, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test public network beam - missing destination_network
	// /api/v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "public")
	urlValues.Add("content_hash", hash)
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/utils/laser/beam", 400, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test public network beam - missing content_hash
	// /api/v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "public")
	urlValues.Add("destination_network", "public")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/utils/laser/beam", 400, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test public network beam - no passphrase
	// /api/v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "public")
	urlValues.Add("destination_network", "public")
	urlValues.Add("content_hash", hash)
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/utils/laser/beam", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test public network beam - with passphrase
	// /api/v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "public")
	urlValues.Add("destination_network", "public")
	urlValues.Add("content_hash", hash)
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/utils/laser/beam", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}
}
