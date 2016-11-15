package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	datestr = "Mon, 02 Jan 2006 15:04:05 -0700"

	baseSite = "https://%s.neocities.org/%s"
	baseAPI  = "https://neocities.org/api/"
	delete   = baseAPI + "delete"
	upload   = baseAPI + "upload"
	list     = baseAPI + "list"

	ua = "neoc client, v1.0"
)

type Item struct {
	Path    string
	IsDir   bool
	Size    uint
	Updated time.Time
}

func Delete(files []string, user, pass string) error {
	val := strings.NewReader(url.Values{"filenames[]": files}.Encode())
	req, err := http.NewRequest(http.MethodPost, delete, val)
	if err != nil {
		return err
	}

	req.SetBasicAuth(user, pass)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if res.StatusCode != 200 {
		io.Copy(os.Stderr, res.Body)
		res.Body.Close()
		return fmt.Errorf("Invalid response (%d)", res.StatusCode)
	}

	return err
}

func Upload(files []string, user, pass string) error {
	var buf bytes.Buffer
	wr := multipart.NewWriter(&buf)
	for _, file := range files {
		var sysf, fname string
		if strings.Contains(file, ":") {
			p := strings.Split(file, ":")
			fname = p[0]
			sysf = p[1]
		} else {
			sysf = file
			fname = file
		}

		fmt.Println(sysf, fname)

		f, err := os.Open(sysf)
		if err != nil {
			return err
		}
		defer f.Close()

		fh, err := wr.CreateFormFile(fname, sysf)
		if err != nil {
			return err
		}

		if _, err = io.Copy(fh, f); err != nil {
			return err
		}
	}
	wr.Close()

	req, err := http.NewRequest(http.MethodPost, upload, &buf)
	if err != nil {
		return err
	}

	req.SetBasicAuth(user, pass)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Content-Type", wr.FormDataContentType())

	if res, err := http.DefaultClient.Do(req); err != nil {
		return err
	} else if res.StatusCode != 200 {
		return fmt.Errorf("Invalid response (%d)", res.StatusCode)
	}

	return err
}

func List(user, pass string) (itms []Item, err error) {
	req, err := http.NewRequest(http.MethodGet, list, nil)
	if err != nil {
		return
	}

	req.SetBasicAuth(user, pass)
	req.Header.Set("User-Agent", ua)

	res, err := http.DefaultClient.Do(req)
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid response (%d)", res.StatusCode)
	} else if err != nil {
		return
	}

	var data struct {
		Result string                   `json:"result"`
		Files  []map[string]interface{} `json:"files"`
	}

	err = json.NewDecoder(res.Body).Decode(&data)
	res.Body.Close()
	if err != nil {
		return
	}

	for _, i := range data.Files {
		date, err := time.Parse(datestr, i["updated_at"].(string))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		size, ok := i["size"].(float64)
		if !ok {
			size = 0
		}

		itms = append(itms, Item{
			Path:    i["path"].(string),
			IsDir:   i["is_directory"].(bool),
			Size:    uint(size),
			Updated: date,
		})
	}

	return
}

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

func Push(dir string, log chan<- string, user, pass string) error {
	datal, err := List(user, pass)
	if err != nil {
		return err
	}
	data := make(map[string]Item)
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
		if !ok || it.Updated.Before(info.ModTime()) {
			log <- rpath + ":" + path
		}

		return nil
	})

	close(log)
	return nil
}
