package neoc

import "time"

const (
	datestr = "Mon, 02 Jan 2006 15:04:05 -0700"

	baseSite = "https://%s.neocities.org/%s"
	baseAPI  = "https://neocities.org/api/"
	delete   = baseAPI + "delete"
	upload   = baseAPI + "upload"
	list     = baseAPI + "list"

	userAgent = "neoc client, v1.0"
)

// Item represents a Document, Image or other file that could be on the
// server
type Item struct {
	Path    string
	IsDir   bool
	Size    uint
	Updated time.Time
}
