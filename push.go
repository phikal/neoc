package neoc

import (
	"os"
	"path/filepath"
)

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
