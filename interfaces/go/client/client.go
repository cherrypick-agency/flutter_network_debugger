package client

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type Client struct {
    BaseURL string
    HTTP    *http.Client
}

func New(baseURL string) *Client { return &Client{BaseURL: baseURL, HTTP: http.DefaultClient} }

type Session struct {
    ID string `json:"id"`
    Target string `json:"target"`
}

func (c *Client) ListSessions(limit, offset int) ([]Session, int, error) {
    req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/sessions?limit=%d&offset=%d", c.BaseURL, limit, offset), nil)
    resp, err := c.HTTP.Do(req)
    if err != nil { return nil, 0, err }
    defer resp.Body.Close()
    var out struct{ Items []Session `json:"items"`; Total int `json:"total"` }
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, 0, err }
    return out.Items, out.Total, nil
}


