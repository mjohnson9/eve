package eve

import (
	"net/http"
	"net/url"
)

// APIBaseURL represents the base URL for all API calls. By default, it is:
// https://api.eveonline.com/
var APIBaseURL = &url.URL{
	Scheme: "https",
	Host:   "api.eveonline.com",
	Path:   "/",
}

// DefaultClient is the default API client.
var DefaultClient = &Client{}

// Call issues a GET request to the specified endpoint, which make be either a
// relative or absolute URL.
func Call(endpoint string, v interface{}) (*Metadata, error) {
	return DefaultClient.Call(endpoint, v)
}

// Client represents an EVE API client. The zero value for Client is a full
// usable API client without any access flags.
type Client struct {
	// HTTPClient specifics which HTTP client is used for making HTTP requests.
	// If nil, http.DefaultClient is used.
	HTTPClient *http.Client
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient == nil {
		return http.DefaultClient
	}

	return c.HTTPClient
}

// Call issues a GET request to the specified endpoint, which make be either a
// relative or absolute URL.
func (c *Client) Call(endpoint string, v interface{}) (*Metadata, error) {
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	resolvedURL := APIBaseURL.ResolveReference(parsedURL)

	res, err := c.httpClient().Get(resolvedURL.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return NewDecoder(res.Body).Decode(v)
}
