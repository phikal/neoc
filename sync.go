package neoc

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"
)

func Sync(dir string, log chan<- string, user, pass string) (int, error) {
	data, err := List(user, pass)
	if err != nil {
		return 0, err
	}

	if dir == "" {
		dir = "."
	} else {
		err = os.MkdirAll(dir, 0755)
		if err != os.ErrExist && err != nil {
			return 0, err
		}
	}

	for _, i := range data {
		go func(item Item) {
			if item.IsDir {
				err = os.MkdirAll(dir+"/"+item.Path, 0755)
				if err != nil && !os.IsExist(os.ErrExist) {
					log <- ""
					return
				}

				err = os.Chtimes(dir+"/"+item.Path, time.Now(), item.Updated)
				if err != nil {
					fmt.Println(err)
				}

				log <- item.Path
				return
			}

			err = os.MkdirAll(dir+"/"+path.Dir(item.Path), 0755)
			if err != nil && !os.IsExist(os.ErrExist) {
				log <- ""
				return
			}

			fs, err := os.Stat(dir + "/" + item.Path)
			if err != nil && !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "%s: %s\n", item.Path, err)
				log <- ""
				return
			}
			if fs != nil && item.Updated.Before(fs.ModTime()) {
				fmt.Fprintf(os.Stderr, "%s: newer than server version, ignoring\n", item.Path)
				log <- ""
				return
			}

			res, err := http.Get(fmt.Sprintf(baseSite, user, item.Path))
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s\n", item.Path, err)
				log <- ""
				return
			}

			f, err := os.Create(dir + "/" + item.Path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s\n", item.Path, err)
				log <- ""
				return
			}
			defer f.Close()

			if _, err := io.Copy(f, res.Body); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s\n", item.Path, err)
				log <- ""
				return
			}

			err = os.Chtimes(f.Name(), time.Now(), item.Updated)
			if err != nil {
				fmt.Println(err)
			}

			log <- item.Path
		}(i)
	}

	return len(data), nil
}
