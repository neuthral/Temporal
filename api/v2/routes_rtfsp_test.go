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
	"github.com/RTradeLtd/database/models"
	pbOrch "github.com/RTradeLtd/grpc/ipfs-orchestrator"
)

func Test_API_Routes_IPFS_Private(t *testing.T) {
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

	nm := models.NewHostedIPFSNetworkManager(db)

	// create private network - failure missing name
	// /api/v2/ipfs/private/new
	var apiResp apiResponse
	urlValues := url.Values{}
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/network/new", 400, nil, nil, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	if apiResp.Code != 400 {
		t.Fatal("bad api status code from /api/v2/ipfs/private/network/new")
	}
	if apiResp.Response != "network_name not present" {
		t.Fatal("failed to detect missing network_name field")
	}

	// create private network - failure name is PUBLIC
	// /api/v2/ipfs/private/new
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("network_name", "PUBLIC")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/network/new", 400, nil, nil, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	if apiResp.Code != 400 {
		t.Fatal("bad api status code from /api/v2/ipfs/private/network/new")
	}

	// create private network - failure name is public
	// /api/v2/ipfs/private/new
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("network_name", "public")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/network/new", 400, nil, nil, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	if apiResp.Code != 400 {
		t.Fatal("bad api status code from /api/v2/ipfs/private/network/new")
	}

	// create private network
	// /api/v2/ipfs/private/new
	var mapAPIResp mapAPIResponse
	urlValues = url.Values{}
	urlValues.Add("network_name", "abc123")
	fakeOrch.StartNetworkReturnsOnCall(0, &pbOrch.StartNetworkResponse{Api: "/ip4/127.0.0.1/tcp/5001", SwarmKey: testSwarmKey}, nil)
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/network/new", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api response status code from /api/v2/ipfs/private/new")
	}
	if mapAPIResp.Response["network_name"] != "abc123" {
		t.Fatal("failed to retrieve correct network name")
	}
	if mapAPIResp.Response["api_url"] != "/ip4/127.0.0.1/tcp/5001" {
		t.Fatal("failed to retrieve correct api url")
	}
	if mapAPIResp.Response["swarm_key"] != testSwarmKey {
		t.Fatal("failed to get correct swarm key")
	}
	if err := nm.UpdateNetworkByName("abc123", map[string]interface{}{
		"api_url": cfg.IPFS.APIConnection.Host + ":" + cfg.IPFS.APIConnection.Port,
	}); err != nil {
		t.Fatal(err)
	}

	// create private network with parameters - invalid bootstrap peer
	// /api/v2/ipfs/private/network
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("network_name", "xyz123")
	urlValues.Add("swarm_key", testSwarmKey)
	urlValues.Add("bootstrap_peers", "not a valid bootstrap peer")
	urlValues.Add("users", "testuser")
	urlValues.Add("users", "testuser2")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/network/new", 400, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}

	// create private network with parameters
	// /api/v2/ipfs/private/network
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("network_name", "xyz123")
	urlValues.Add("swarm_key", testSwarmKey)
	urlValues.Add("bootstrap_peers", testBootstrapPeer1)
	urlValues.Add("bootstrap_peers", testBootstrapPeer2)
	urlValues.Add("users", "testuser")
	urlValues.Add("users", "testuser2")
	fakeOrch.StartNetworkReturnsOnCall(1, &pbOrch.StartNetworkResponse{Api: "/ip4/127.0.0.1/tcp/5002", SwarmKey: "swarmStorm"}, nil)
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/network/new", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api response status code from /api/v2/ipfs/private/new")
	}
	if mapAPIResp.Response["network_name"] != "xyz123" {
		t.Fatal("failed to retrieve correct network name")
	}
	if mapAPIResp.Response["api_url"] != "/ip4/127.0.0.1/tcp/5002" {
		t.Fatal("failed to retrieve correct api url")
	}
	if mapAPIResp.Response["swarm_key"] != "swarmStorm" {
		t.Fatal("failed to get correct swarm key")
	}

	// get private network information
	// /api/v2/ipfs/private/network/:name
	var interfaceAPIResp interfaceAPIResponse
	if err := sendRequest(
		api, "GET", "/api/v2/ipfs/private/network/abc123", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad api response status code from /api/v2/ipfs/private/network/abc123")
	}

	// get all authorized private networks
	// /api/v2/ipfs/private/networks
	var stringSliceAPIResp stringSliceAPIResponse
	if err := sendRequest(
		api, "GET", "/api/v2/ipfs/private/networks", 200, nil, nil, &stringSliceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if stringSliceAPIResp.Code != 200 {
		t.Fatal("bad api response status code from /api/v2/ipfs/private/networks")
	}
	if len(stringSliceAPIResp.Response) == 0 {
		t.Fatal("failed to find any from /api/v2/ipfs/private/networks")
	}
	var found bool
	for _, v := range stringSliceAPIResp.Response {
		if v == "abc123" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("failed to find correct network from /api/v2/ipfs/private/networks")
	}

	// stop private network - missing network_name
	// /api/v2/ipfs/private/network/stop
	mapAPIResp = mapAPIResponse{}
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/network/stop", 400, nil, nil, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api resposne code from /api/v2/ipfs/private/network/stop")
	}
	if mapAPIResp.Response["state"] != "stopped" {
		t.Fatal("failed to stop network")
	}

	// stop private network - invalid network access
	// /api/v2/ipfs/private/network/stop
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("network_name", "abc123somerandomnetworknotownedbyuuuuuuuuu")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/network/stop", 400, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}

	// stop private network
	// /api/v2/ipfs/private/network/stop
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("network_name", "abc123")
	fakeOrch.StopNetworkReturnsOnCall(0, &pbOrch.Empty{}, nil)
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/network/stop", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api resposne code from /api/v2/ipfs/private/network/stop")
	}
	if mapAPIResp.Response["state"] != "stopped" {
		t.Fatal("failed to stop network")
	}

	// start private network - missing network name
	// /api/v2/ipfs/private/network/start
	mapAPIResp = mapAPIResponse{}
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/network/start", 400, nil, nil, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}

	// start private network - invalid network access
	// /api/v2/ipfs/private/network/start
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("network_name", "abc123somerandomnetworknotownedbyuuuuuuuuu")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/network/start", 400, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}

	// start private network
	// /api/v2/ipfs/private/network/start
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("network_name", "abc123")
	fakeOrch.StartNetworkReturnsOnCall(2, &pbOrch.StartNetworkResponse{Api: "test", SwarmKey: "test"}, nil)
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/network/start", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api resposne code from /api/v2/ipfs/private/network/stop")
	}
	if mapAPIResp.Response["state"] != "started" {
		t.Fatal("failed to stop network")
	}

	// add a file normally
	// /api/v2/ipfs/private/file/add
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
	req := httptest.NewRequest("POST", "/api/v2/ipfs/private/file/add", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	urlValues.Add("network_name", "abc123")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /api/v2/ipfs/private/file/add")
	}
	apiResp = apiResponse{}
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
		t.Fatal("bad api status code from /api/v2/ipfs/private/file/add")
	}
	hash = apiResp.Response

	// add a file advanced
	// /api/v2/ipfs/private/file/add/advanced
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
	req = httptest.NewRequest("POST", "/api/v2/ipfs/private/file/add/advanced", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	urlValues.Add("passphrase", "password123")
	urlValues.Add("network_name", "abc123")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /api/v2/ipfs/private/file/add/advanced")
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
		t.Fatal("bad api status code from /api/v2/ipfs/private/file/add/advanced")
	}

	// test pinning - missing hold_time
	// /api/v2/ipfs/private/pin
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("network_name", "abc123")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/pin/"+hash, 400, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}

	// test pinning - missing network_name
	// /api/v2/ipfs/private/pin
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/pin/"+hash, 400, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}

	// test pinning - invalid network access error
	// /api/v2/ipfs/private/pin
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	urlValues.Add("network_name", "abc123somerandomnetworknotownedbyuuuuuuuuu")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/pin/"+hash, 400, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}

	// test pinning
	// /api/v2/ipfs/private/pin
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	urlValues.Add("network_name", "abc123")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/pin/"+hash, 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from  /api/v2/ipfs/private/pin")
	}

	// test pin check - invalid network access
	// /api/v2/ipfs/private/check/pin
	var boolAPIResp boolAPIResponse
	if err := sendRequest(
		api, "GET", "/api/v2/ipfs/private/pin/check/"+hash+"/abc123", 200, nil, nil, &boolAPIResp,
	); err != nil {
		t.Fatal(err)
	}

	// test pin check
	// /api/v2/ipfs/private/check/pin
	boolAPIResp = boolAPIResponse{}
	if err := sendRequest(
		api, "GET", "/api/v2/ipfs/private/pin/check/"+hash+"/abc123notarealnettwrrook", 400, nil, nil, &boolAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if boolAPIResp.Code != 200 {
		t.Fatal("bad api status code from  /api/v2/ipfs/private/check/pin")
	}

	// test pubsub publish - missing message
	// /api/v2/ipfs/private/publish/topic
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("network_name", "abc123")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/pubsub/publish/foo", 400, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}

	// test pubsub publish - missing network
	// /api/v2/ipfs/private/publish/topic
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("message", "bar")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/pubsub/publish/foo", 400, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}

	// test pubsub publish - invalid network access error
	// /api/v2/ipfs/private/publish/topic
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("message", "bar")
	urlValues.Add("network_name", "abc123somerandomnetworknotownedbyuuuuuuuuu")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/pubsub/publish/foo", 400, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}

	// test pubsub publish
	// /api/v2/ipfs/private/publish/topic
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("message", "bar")
	urlValues.Add("network_name", "abc123")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/private/pubsub/publish/foo", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api status code from  /api/v2/ipfs/private/pubsub/publish/topic")
	}
	if mapAPIResp.Response["topic"] != "foo" {
		t.Fatal("bad response")
	}
	if mapAPIResp.Response["message"] != "bar" {
		t.Fatal("bad response")
	}

	// test object stat - invalid network access error
	// /api/v2/ipfs/private/stat
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/api/v2/ipfs/private/stat/"+hash+"/abc123lazytotype", 400, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/private/stat")
	}

	// test object stat
	// /api/v2/ipfs/private/stat
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/api/v2/ipfs/private/stat/"+hash+"/abc123", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/private/stat")
	}

	// test get dag - invalid network access
	// /api/v2/ipfs/private/dag
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/api/v2/ipfs/private/dag/"+hash+"/abc123lazy", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}

	// test get dag
	// /api/v2/ipfs/private/dag
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/api/v2/ipfs/private/dag/"+hash+"/abc123", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/private/dag/")
	}
	// test download - invalid netwokr access error
	// /api/v2/ipfs/utils/download
	urlValues = url.Values{}
	urlValues.Add("network_name", "screwdriversaregooddrinksonplanes")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/utils/download/"+hash, 400, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test download
	// /api/v2/ipfs/utils/download
	urlValues = url.Values{}
	urlValues.Add("network_name", "abc123")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/utils/download/"+hash, 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test get authorized networks
	// /api/v2/ipfs/private/networks
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/api/v2/ipfs/private/networks", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/private/networks/")
	}

	// test get authorized networks
	// /api/v2/ipfs/private/networks
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/api/v2/ipfs/private/uploads/abc123", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/private/uploads")
	}

	// test private network beam - source private, dest public
	urlValues = url.Values{}
	urlValues.Add("source_network", "abc123")
	urlValues.Add("destination_network", "public")
	urlValues.Add("content_hash", hash)
	urlValues.Add("passphrase", "password123")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/utils/laser/beam", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test private network beam - source public, dest private
	// /api/v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "public")
	urlValues.Add("destination_network", "abc123")
	urlValues.Add("content_hash", hash)
	urlValues.Add("passphrase", "password123")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/utils/laser/beam", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test private network beam - source private, dest private
	// /api/v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "abc123")
	urlValues.Add("destination_network", "abc123")
	urlValues.Add("content_hash", hash)
	urlValues.Add("passphrase", "password123")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/utils/laser/beam", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test private network beam - invalid network access error (source)
	// /api/v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "readytoland")
	urlValues.Add("destination_network", "abc123")
	urlValues.Add("content_hash", hash)
	urlValues.Add("passphrase", "password123")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/utils/laser/beam", 400, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test private network beam - invalid network access error (dest)
	// /api/v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "abc123")
	urlValues.Add("destination_network", "planesarefastnowadays")
	urlValues.Add("content_hash", hash)
	urlValues.Add("passphrase", "password123")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/utils/laser/beam", 400, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// remove private network - missing network_name
	// /api/v2/ipfs/private/network/remove
	mapAPIResp = mapAPIResponse{}
	if err := sendRequest(
		api, "DELETE", "/api/v2/ipfs/private/network/remove", 400, nil, nil, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}

	// remove private network - invalid network access
	// /api/v2/ipfs/private/network/remove
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("network_name", "900mkh")
	if err := sendRequest(
		api, "DELETE", "/api/v2/ipfs/private/network/remove", 400, nil, nil, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}

	// remove private network
	// /api/v2/ipfs/private/network/remove
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("network_name", "abc123")
	fakeOrch.RemoveNetworkReturnsOnCall(0, &pbOrch.Empty{}, nil)
	if err := sendRequest(
		api, "DELETE", "/api/v2/ipfs/private/network/remove", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api response status code from /api/v2/ipfs/private/network/remove")
	}
	if mapAPIResp.Response["state"] != "removed" {
		t.Fatal("failed to remove network")
	}
}
