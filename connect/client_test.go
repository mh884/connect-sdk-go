package connect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
)

var validHost string
var validToken string
var defaultVault string

var mockHTTPClient *mockClient
var testClient *restClient

var requestCount int
var requestFail bool

type mockClient struct {
	Dofunc func(req *http.Request) (*http.Response, error)
}

func (mc *mockClient) Do(req *http.Request) (*http.Response, error) {
	return mc.Dofunc(req)
}

func TestMain(m *testing.M) {
	validHost = "http://localhost:8080"
	validToken = "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyIxcGFzc3dvcmQuY29tL2F1dWlkIjoiR1RLVjVWUk5UUkVHUEVMWE41QlBSQTJHTjQiLCIxcGFzc3dvcmQuY29tL2Z0cyI6WyJ2YXVsdGFjY2VzcyJdLCIxcGFzc3dvcmQuY29tL3Rva2VuIjoidTFxMUNtWVhtbGR2YWZUa0lHTW8tLTJnazl5a180SkMiLCIxcGFzc3dvcmQuY29tL3Z0cyI6W3siYSI6MTU3MzA2NzIsInUiOiJvdGw2cjZudWdqNXdyNjNybmt3M3Y0cGJuYSJ9XSwiYXVkIjpbImNvbS4xcGFzc3dvcmQuc2VjcmV0c2VydmljZSJdLCJpYXQiOjE2MDMxMjg2NDIsImlzcyI6ImNvbS4xcGFzc3dvcmQuYjUiLCJqdGkiOiI2bjYyZHhyanBxZW00aGJ4d3dxdGJtNmpsZSIsInN1YiI6IkFWNFFORUM3UFJGREZFRTJJREpNM0NSSUNJIn0.-1shD95-qGYrHh3beH5nrfsV91BMp30Y9ipIwE6n4pw8Y9-2fR-gun27ShS9fHLJqW9xJZ-Eir1UEkiha2ucvA"
	defaultVault = "otl6r6nugj5wr63rnkw3v4pbna"

	os.Setenv("OP_VAULT", defaultVault)
	os.Setenv("OP_CONNECT_HOST", validHost)
	os.Setenv("OP_CONNECT_TOKEN", validToken)

	mockHTTPClient = &mockClient{}

	testClient = &restClient{
		URL:       validHost,
		Token:     validToken,
		userAgent: defaultUserAgent,
		tracer:    opentracing.GlobalTracer(),
		client:    mockHTTPClient,
	}

	requestCount = 0
	requestFail = false

	os.Exit(m.Run())
}

func TestNewClientFromEnvironmentWithoutHost(t *testing.T) {
	os.Unsetenv("OP_CONNECT_HOST")
	defer os.Setenv("OP_CONNECT_HOST", validHost)
	_, err := NewClientFromEnvironment()
	if err == nil {
		t.Log("Expected client to fail")
		t.FailNow()
	}
}

func TestNewClientFromEnvironmentWithoutToken(t *testing.T) {
	os.Unsetenv("OP_CONNECT_TOKEN")
	defer os.Setenv("OP_CONNECT_TOKEN", validToken)
	_, err := NewClientFromEnvironment()
	if err == nil {
		t.Log("Expected client to fail")
		t.FailNow()
	}
}

func TestNewClientFromEnvironment(t *testing.T) {
	client, err := NewClientFromEnvironment()
	if err != nil {
		t.Logf("Unable to create client from environment: %q", err)
		t.FailNow()
	}

	restClient, ok := client.(*restClient)
	if !ok {
		t.Log("Unable to cast client to rest client. Was expecting restClient")
		t.FailNow()
	}

	if restClient.URL != validHost {
		t.Logf("Expected host of http://localhost:8080, got %q", restClient.URL)
		t.FailNow()
	}

	if restClient.Token != validToken {
		t.Logf("Expected valid token %q, got %q", validToken, restClient.Token)
		t.FailNow()
	}

	if restClient.userAgent != defaultUserAgent {
		t.Logf("Expected user-agent of %q, got %q", defaultUserAgent, restClient.userAgent)
		t.FailNow()
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient(validHost, validToken)

	restClient, ok := client.(*restClient)
	if !ok {
		t.Log("Unable to cast client to rest client. Was expecting restClient")
		t.FailNow()
	}

	if restClient.URL != validHost {
		t.Logf("Expected host of http://localhost:8080, got %q", restClient.URL)
		t.FailNow()
	}

	if restClient.Token != validToken {
		t.Logf("Expected valid token %q, got %q", validToken, restClient.Token)
		t.FailNow()
	}

	if restClient.userAgent != defaultUserAgent {
		t.Logf("Expected user-agent of %q, got %q", defaultUserAgent, restClient.userAgent)
		t.FailNow()
	}
}

func TestNewClientWithUserAgent(t *testing.T) {
	client := NewClientWithUserAgent(validHost, validToken, "testSuite")

	restClient, ok := client.(*restClient)
	if !ok {
		t.Log("Unable to cast client to rest client. Was expecting restClient")
		t.FailNow()
	}

	if restClient.URL != validHost {
		t.Logf("Expected host of http://localhost:8080, got %q", restClient.URL)
		t.FailNow()
	}

	if restClient.Token != validToken {
		t.Logf("Expected valid token %q, got %q", validToken, restClient.Token)
		t.FailNow()
	}

	if restClient.userAgent != "testSuite" {
		t.Logf("Expected user-agent of %q, got %q", defaultUserAgent, restClient.userAgent)
		t.FailNow()
	}

}

func Test_restClient_GetVaults(t *testing.T) {
	mockHTTPClient.Dofunc = listVaults
	vaults, err := testClient.GetVaults()

	if err != nil {
		t.Logf("Unable to get vaults: %s", err.Error())
		t.FailNow()
	}

	if len(vaults) < 1 {
		t.Logf("Expected vaults to exist, found %d", len(vaults))
		t.FailNow()
	}
}

func Test_restClient_GetVaultsByTitle(t *testing.T) {
	mockHTTPClient.Dofunc = listVaults
	vaults, err := testClient.GetVaultsByTitle("Test Vault")

	if err != nil {
		t.Logf("Unable to get vaults: %s", err.Error())
		t.FailNow()
	}

	if len(vaults) < 1 {
		t.Logf("Expected vaults to exist, found %d", len(vaults))
		t.FailNow()
	}
}

func Test_restClient_GetItem(t *testing.T) {
	mockHTTPClient.Dofunc = getItem
	item, err := testClient.GetItem(uuid.New().String(), uuid.New().String())

	if err != nil {
		t.Logf("Unable to get items: %s", err.Error())
		t.FailNow()
	}

	if item == nil {
		t.Log("Expected 1 item to exist")
		t.FailNow()
	}
}

func Test_restClient_GetItemNotFound(t *testing.T) {
	requestFail = true
	defer reset()

	mockHTTPClient.Dofunc = getItem
	item, err := testClient.GetItem(uuid.New().String(), uuid.New().String())

	if err == nil {
		t.Log("Expected a 404")
		t.FailNow()
	}

	if item != nil {
		t.Log("Expected no items returns")
		t.FailNow()
	}
}

func Test_restClient_GetItems(t *testing.T) {
	mockHTTPClient.Dofunc = listItems
	items, err := testClient.GetItems(uuid.New().String())

	if err != nil {
		t.Logf("Unable to get item: %s", err.Error())
		t.FailNow()
	}

	if len(items) != 1 {
		t.Logf("Expected 1 item to exist in vault, found %d", len(items))
		t.FailNow()
	}
}

func Test_restClient_GetItemsByTitle(t *testing.T) {
	mockHTTPClient.Dofunc = listItems
	items, err := testClient.GetItemsByTitle("test", uuid.New().String())

	if err != nil {
		t.Logf("Unable to get item: %s", err.Error())
		t.FailNow()
	}

	if len(items) != 1 {
		t.Logf("Expected 1 item to exist in vault, found %d", len(items))
		t.FailNow()
	}
}

func Test_restClient_GetItemByTitle(t *testing.T) {
	defer reset()

	mockHTTPClient.Dofunc = getItemByID
	item, err := testClient.GetItemByTitle("test", uuid.New().String())

	if err != nil {
		t.Logf("Unable to get item: %s", err.Error())
		t.FailNow()
	}

	if item == nil {
		t.Log("Expected 1 item to exist")
		t.FailNow()
	}
}

func Test_restClient_GetItemByNonUniqueTitle(t *testing.T) {
	requestFail = true
	defer reset()

	mockHTTPClient.Dofunc = getItemByID
	item, err := testClient.GetItemByTitle("test", uuid.New().String())

	if err == nil {
		t.Log("Expected too many items")
		t.FailNow()
	}

	if item != nil {
		t.Log("Expected no items returns")
		t.FailNow()
	}
}

func Test_restClient_CreateItem(t *testing.T) {
	mockHTTPClient.Dofunc = createItem
	item, err := testClient.CreateItem(generateItem(defaultVault), defaultVault)

	if err != nil {
		t.Logf("Unable to create items: %s", err.Error())
		t.FailNow()
	}

	if item == nil {
		t.Log("Expected 1 item to be created")
		t.FailNow()
	}
}

func Test_restClient_CreateItemError(t *testing.T) {
	requestFail = true
	defer reset()

	mockHTTPClient.Dofunc = createItem
	item, err := testClient.CreateItem(generateItem(defaultVault), defaultVault)

	if err == nil {
		t.Log("Should have failed to create item")
		t.FailNow()
	}

	if item != nil {
		t.Log("Expected item to not be created")
		t.FailNow()
	}
}

func Test_restClient_UpdateItem(t *testing.T) {
	mockHTTPClient.Dofunc = updateItem
	item, err := testClient.UpdateItem(generateItem(defaultVault), defaultVault)

	if err != nil {
		t.Logf("Unable to update item: %s", err.Error())
		t.FailNow()
	}

	if item == nil {
		t.Log("Expected 1 item to be updated")
		t.FailNow()
	}
}

func Test_restClient_UpdateItemError(t *testing.T) {
	requestFail = true
	defer reset()

	mockHTTPClient.Dofunc = updateItem
	item, err := testClient.UpdateItem(generateItem(defaultVault), defaultVault)

	if err == nil {
		t.Log("Should have failed to update item")
		t.FailNow()
	}

	if item != nil {
		t.Log("Expected item to not be update")
		t.FailNow()
	}
}

func Test_restClient_DeleteItem(t *testing.T) {
	mockHTTPClient.Dofunc = deleteItem
	err := testClient.DeleteItem(generateItem(defaultVault), defaultVault)

	if err != nil {
		t.Logf("Unable to delete item: %s", err.Error())
		t.FailNow()
	}
}

func Test_restClient_DeleteItemError(t *testing.T) {
	requestFail = true
	defer reset()

	mockHTTPClient.Dofunc = deleteItem
	err := testClient.DeleteItem(generateItem(defaultVault), defaultVault)

	if err == nil {
		t.Logf("Expected delete to fail")
		t.FailNow()
	}
}

func listVaults(req *http.Request) (*http.Response, error) {
	vaults := []onepassword.Vault{
		{
			Description: "Test Vault",
			ID:          uuid.New().String(),
		},
	}

	json, _ := json.Marshal(vaults)
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(json)),
		Header:     req.Header,
	}, nil
}

func generateItem(vaultUUID string) *onepassword.Item {
	return &onepassword.Item{
		ID: uuid.New().String(),
		Vault: onepassword.ItemVault{
			ID: vaultUUID,
		},
	}
}

func listItems(req *http.Request) (*http.Response, error) {
	vaultUUID := ""
	excessPath := ""
	fmt.Sscanf(req.URL.Path, "/v1/vaults/%s%s", vaultUUID, excessPath)

	items := []*onepassword.Item{
		generateItem(vaultUUID),
	}

	json, _ := json.Marshal(items)
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(json)),
		Header:     req.Header,
	}, nil
}

func getItemByID(req *http.Request) (*http.Response, error) {
	vaultUUID := ""
	excessPath := ""
	fmt.Sscanf(req.URL.Path, "/v1/vaults/%s%s", vaultUUID, excessPath)

	items := []*onepassword.Item{
		generateItem(vaultUUID),
	}

	if requestFail {
		items = append(items, generateItem(vaultUUID))
	}

	if requestCount == 0 {
		requestCount++
		json, _ := json.Marshal(items)
		return &http.Response{
			Status:     http.StatusText(http.StatusOK),
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(json)),
			Header:     req.Header,
		}, nil
	}

	return getItem(req)
}

func getItem(req *http.Request) (*http.Response, error) {
	if requestFail {
		json, _ := json.Marshal("Not found")
		return &http.Response{
			Status:     http.StatusText(http.StatusNotFound),
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader(json)),
			Header:     req.Header,
		}, nil
	}

	vaultUUID := ""
	excessPath := ""
	fmt.Sscanf(req.URL.Path, "/v1/vaults/%s%s", vaultUUID, excessPath)

	json, _ := json.Marshal(generateItem(vaultUUID))
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(json)),
		Header:     req.Header,
	}, nil
}

func createItem(req *http.Request) (*http.Response, error) {
	if requestFail {
		json, _ := json.Marshal("Vault UUID required")
		return &http.Response{
			Status:     http.StatusText(http.StatusBadRequest),
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(bytes.NewReader(json)),
			Header:     req.Header,
		}, nil
	}

	rawBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var item onepassword.Item
	if err := json.Unmarshal(rawBody, &item); err != nil {
		return nil, err
	}

	newUUID := uuid.New()
	item.ID = newUUID.String()
	item.CreatedAt = time.Now()

	vaultUUID := ""
	excessPath := ""
	fmt.Sscanf(req.URL.Path, "/v1/vaults/%s%s", vaultUUID, excessPath)

	item.Vault.ID = vaultUUID

	item.UpdatedAt = time.Now()

	json, _ := json.Marshal(item)
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(json)),
		Header:     req.Header,
	}, nil
}

func updateItem(req *http.Request) (*http.Response, error) {
	if requestFail {
		json, _ := json.Marshal("Missing required field")
		return &http.Response{
			Status:     http.StatusText(http.StatusBadRequest),
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(bytes.NewReader(json)),
			Header:     req.Header,
		}, nil
	}

	rawBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var item onepassword.Item
	if err := json.Unmarshal(rawBody, &item); err != nil {
		return nil, err
	}

	item.UpdatedAt = time.Now()

	json, _ := json.Marshal(item)
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(json)),
		Header:     req.Header,
	}, nil
}

func deleteItem(req *http.Request) (*http.Response, error) {
	if requestFail {
		json, _ := json.Marshal("Vault not found")
		return &http.Response{
			Status:     http.StatusText(http.StatusNotFound),
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader(json)),
			Header:     req.Header,
		}, nil
	}

	vaultUUID := ""
	itemUUID := ""
	fmt.Sscanf(req.URL.Path, "/v1/vaults/%s/items/%s", vaultUUID, itemUUID)

	return &http.Response{
		Status:     http.StatusText(http.StatusNoContent),
		StatusCode: http.StatusNoContent,
		Header:     req.Header,
	}, nil
}

func reset() {
	requestCount = 0
	requestFail = false
}