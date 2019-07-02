package neoc

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"
)

func (c *Client) upload(dir string, item *Item, log chan<- *Item) {
	var err error

	if item.IsDir {
		err = os.MkdirAll(dir+"/"+item.Path, 0755)
		if err != nil && !os.IsExist(os.ErrExist) {
			log <- item
			return
		}

		err = os.Chtimes(dir+"/"+item.Path, time.Now(), item.Updated)
		if err != nil {
			fmt.Println(err)
		}

		log <- item
		return
	}

	err = os.MkdirAll(dir+"/"+path.Dir(item.Path), 0755)
	if err != nil && !os.IsExist(os.ErrExist) {
		log <- nil
		return
	}

	fs, err := os.Stat(dir + "/" + item.Path)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%s: %s\n", item.Path, err)
		log <- nil
		return
	}
	if fs != nil && item.Updated.Before(fs.ModTime()) {
		fmt.Fprintf(os.Stderr, "%s: newer than server version, ignoring\n", item.Path)
		log <- nil
		return
	}

	res, err := http.Get(fmt.Sprintf(c.base, c.user, item.Path))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", item.Path, err)
		log <- nil
		return
	}

	f, err := os.Create(dir + "/" + item.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", item.Path, err)
		log <- nil
		return
	}
	defer f.Close()

	if _, err := io.Copy(f, res.Body); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", item.Path, err)
		log <- nil
		return
	}

	err = os.Chtimes(f.Name(), time.Now(), item.Updated)
	if err != nil {
		fmt.Println(err)
	}

	log <- item
}

// Sync will attempt to update the client version of the directory dir
// on the server, and vice versa. The method will block until it is
// finished.
//
// The channel log will report the progress if log is not nil.
func (c *Client) Sync(dir string, log chan<- *Item) (int, error) {
	data, err := c.List()
	if err != nil {
		return 0, err
	}

	err = os.MkdirAll(dir, 0755)
	if err != os.ErrExist && err != nil {
		return 0, err
	}

	ilog := make(chan *Item)
	for _, i := range data {
		go c.upload(dir, i, ilog)
	}

	for i := range ilog {
		if log != nil {
			log <- i
		}
	}
	return len(data), nil
}
