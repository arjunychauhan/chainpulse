package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Client struct {
	URL        string
	HTTPClient *http.Client
}
type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *RPCError       `json:"error"`
}

func NewClient(url string) *Client {
	return &Client{
		URL:        url,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) Call(ctx context.Context, method string) ([]byte, error) {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  []interface{}{},
		ID:      1,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.URL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) callUint64(ctx context.Context, method string) (uint64, error) {
	data, err := c.Call(ctx, method)

	if err != nil {
		return 0, err
	}
	var response RPCResponse
	err = json.Unmarshal(data, &response)

	if err != nil {
		return 0, err
	}

	if response.Error != nil {
		return 0, fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	var resultHex string

	err = json.Unmarshal(response.Result, &resultHex)
	if err != nil {
		return 0, err
	}

	result, err := strconv.ParseUint(resultHex, 0, 64)

	if err != nil {
		return 0, err
	}

	return result, nil
}

func (c *Client) ChainID(ctx context.Context) (uint64, error) {
	return c.callUint64(ctx, "eth_chainId")
}

func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	return c.callUint64(ctx, "eth_blockNumber")

}
