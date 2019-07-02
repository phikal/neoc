package neoc

import (
	"os"
	"path/filepath"
)

// Push will send all files from a directory to the client's account,
// unless a newer version is found on the server
//
// If the second argument, log, is not nil, it will report which items are
// being uploaded.
func (c *Client) Push(dir string, log chan<- *Item) error {
	datal, err := c.List()
	if err != nil {
		return err
	}
	data := make(map[string]*Item)
	for _, d := range datal {
		data[d.Path] = d
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		rpath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		it, ok := data[rpath] // ok == false => new file
		if !ok || log == nil || it.Updated.Before(info.ModTime()) {
			log <- it
		}

		return nil
	})

	if log != nil {
		close(log)
	}
	return nil
}
