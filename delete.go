package neoc

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

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
