package neoc

// Client represents a
type Client struct {
	user, pass string
	base, api  string
}

// NewClient creates a new client
func NewClient(user, pass string) *Client {
	return &Client{
		user: user,
		pass: pass,
		api:  "https://neocities.org/api/",
		base: "https://%s.neocities.org/%s",
	}
}
