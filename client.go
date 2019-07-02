package neoc

const (
	defaultAPI  = "https://neocities.org/api/"
	defaultBase = "https://%s.neocities.org/%s"
)

// Client represents a
type Client struct {
	user, pass string
	base, api  string
}

// NewClient creates a new client
func NewClient(user, pass string) *Client {
	return NewClientWithAPI(user, pass, defaultAPI, defaultBase)
}

// NewClientWithAPI creates a client with custom API/Base URL
func NewClientWithAPI(user, pass, api, base string) *Client {
	return &Client{
		user: user,
		pass: pass,
		api:  api,
		base: base,
	}
}
