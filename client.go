package open_wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	client *http.Client
}

func NewClient(cli *http.Client) *Client {
	if cli == nil {
		cli = http.DefaultClient
	}
	return &Client{cli}
}

// post  表单请求
func (cli *Client) PostJson(incompleteURL string, request interface{}, response interface{}) error {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(&request); err != nil {
		return err
	}
	httpResp, err := cli.client.Post(incompleteURL, "application/json; charset=utf-8", &buf)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("http.Status: %s", httpResp.Status)
	}
	return json.NewDecoder(httpResp.Body).Decode(&response)
}
