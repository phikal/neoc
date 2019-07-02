package neoc

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

func (c *Client) Upload(files []string) error {
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

	req, err := http.NewRequest(http.MethodPost, c.api+"/upload", &buf)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.user, c.pass)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", wr.FormDataContentType())

	if res, err := http.DefaultClient.Do(req); err != nil {
		return err
	} else if res.StatusCode != 200 {
		return fmt.Errorf("Invalid response (%d)", res.StatusCode)
	}

	return err
}
