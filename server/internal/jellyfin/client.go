package jellyfin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client talks to a Jellyfin server. It only ever calls /Sessions and the image
// endpoint; both go through the same HTTP client.
type Client struct {
	base  string
	token string
	hc    *http.Client
}

func NewClient(base, token string) *Client {
	return &Client{
		base:  strings.TrimRight(base, "/"),
		token: token,
		hc:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) auth(r *http.Request) {
	r.Header.Set("Authorization", fmt.Sprintf(
		`MediaBrowser Token="%s", Client="Marquee", Device="Marquee", DeviceId="marquee", Version="1.0.0"`,
		c.token))
	r.Header.Set("Accept", "application/json")
}

// Sessions fetches the active session list.
func (c *Client) Sessions(ctx context.Context) ([]Session, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/Sessions", nil)
	if err != nil {
		return nil, err
	}
	c.auth(req)

	res, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4<<10))
		return nil, &HTTPError{Status: res.StatusCode}
	}

	var out []Session
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode sessions: %w", err)
	}
	return out, nil
}

// Image fetches an item's primary image, scaled by Jellyfin to maxWidth.
// tag is forwarded so Jellyfin can serve a cached render.
func (c *Client) Image(ctx context.Context, id, tag string, maxWidth int) (*http.Response, error) {
	target := fmt.Sprintf("%s/Items/%s/Images/Primary?maxWidth=%d", c.base, url.PathEscape(id), maxWidth)
	if tag != "" {
		target += "&tag=" + url.QueryEscape(tag)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	c.auth(req)

	res, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, &HTTPError{Status: res.StatusCode}
	}
	return res, nil
}
