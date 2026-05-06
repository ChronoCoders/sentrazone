package wgeasy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
)

type WGClient struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Enabled           bool    `json:"enabled"`
	Address           string  `json:"address"`
	PublicKey         string  `json:"publicKey"`
	LatestHandshakeAt *string `json:"latestHandshakeAt"`
	TransferRx        int64   `json:"transferRx"`
	TransferTx        int64   `json:"transferTx"`
}

type Client struct {
	baseURL  string
	password string
	hc       *http.Client
	mu       sync.Mutex
	authed   bool
}

var ServerHosts = map[string]string{
	"los-angeles": "http://wg-easy:51821",
	"virginia":    "http://wg-easy-va:51821",
	"dallas":      "http://wg-easy-tx:51821",
}

func New(baseURL, password string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL:  baseURL,
		password: password,
		hc:       &http.Client{Jar: jar},
	}, nil
}

func (c *Client) login(ctx context.Context) error {
	body, _ := json.Marshal(map[string]string{"password": c.password})
	req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/session", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("wg-easy auth failed: %d", resp.StatusCode)
	}
	c.authed = true
	return nil
}

func (c *Client) do(ctx context.Context, method, path string, bodyBytes []byte) (*http.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.authed {
		if err := c.login(ctx); err != nil {
			return nil, err
		}
	}

	makeReq := func() (*http.Request, error) {
		var body io.Reader
		if bodyBytes != nil {
			body = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
		if err == nil && bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		return req, err
	}

	req, err := makeReq()
	if err != nil {
		return nil, err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		resp.Body.Close()
		c.authed = false
		if err := c.login(ctx); err != nil {
			return nil, err
		}
		req, err = makeReq()
		if err != nil {
			return nil, err
		}
		return c.hc.Do(req)
	}
	return resp, nil
}

func (c *Client) CreateClient(ctx context.Context, name string) (*WGClient, error) {
	body, _ := json.Marshal(map[string]string{"name": name})
	resp, err := c.do(ctx, "POST", "/api/wireguard/client", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var client WGClient
	return &client, json.NewDecoder(resp.Body).Decode(&client)
}

func (c *Client) GetClients(ctx context.Context) ([]WGClient, error) {
	resp, err := c.do(ctx, "GET", "/api/wireguard/client", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var clients []WGClient
	return clients, json.NewDecoder(resp.Body).Decode(&clients)
}

func (c *Client) SetEnabled(ctx context.Context, clientID string, enabled bool) error {
	action := "enable"
	if !enabled {
		action = "disable"
	}
	path := fmt.Sprintf("/api/wireguard/client/%s/%s", url.PathEscape(clientID), action)
	resp, err := c.do(ctx, "POST", path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) GetQRCode(ctx context.Context, clientID string) ([]byte, error) {
	path := fmt.Sprintf("/api/wireguard/client/%s/qrcode.svg", url.PathEscape(clientID))
	resp, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
