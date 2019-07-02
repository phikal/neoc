package neoc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// List returns a list of all files/directories as items
func (c *Client) List() (itms []*Item, err error) {
	req, err := http.NewRequest(http.MethodGet, c.api+"/list", nil)
	if err != nil {
		return
	}

	req.SetBasicAuth(c.user, c.pass)
	req.Header.Set("User-Agent", userAgent)

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

		itms = append(itms, &Item{
			Path:    i["path"].(string),
			IsDir:   i["is_directory"].(bool),
			Size:    uint(size),
			Updated: date,
		})
	}

	return
}
