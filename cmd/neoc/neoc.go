package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/phikal/neoc"
)

var verbose bool

func do(c *neoc.Client, mode byte) error {
	switch mode {
	case 'u': // upload
		if flag.NArg() < 2 {
			fmt.Printf("UPLOAD request requires arguments\n")
			os.Exit(1)
		}

		if verbose {
			fmt.Println("Uploading...")
		}
		return c.Upload(flag.Args()[1:])
	case 'd': // delete
		if flag.NArg() < 2 {
			fmt.Printf("DELETE request requires arguments\n")
			os.Exit(1)
		}

		if verbose {
			fmt.Println("Deleting...")
		}
		return c.Delete(flag.Args()[1:])
	case 's': // sync
		ch := make(chan *neoc.Item)
		size, err := c.Sync(flag.Arg(1), ch)
		if err != nil {
			return err
		}

		if verbose {
			fmt.Println("Downloading...")
		}

		i := 1
		for item := range ch {
			if item != nil {
				continue
			}

			if verbose {
				fmt.Printf("[%d/%d] %s...\n", i, size, item.Path)
				i++
			}
		}
	case 'p': // push
		ch := make(chan *neoc.Item)
		err := c.Push(flag.Arg(1), ch)
		if err != nil {
			return err
		}

		if verbose {
			fmt.Println("Pushing...")
		}

		var items []*neoc.Item
		for msg := range ch {
			items = append(items, msg)
			if verbose {
				fmt.Println(msg)
			}
		}

		var files []string
		for _, i := range items {
			files = append(files, i.Path)
		}
		return c.Upload(files)
	case 'l': // list
		if verbose {
			fmt.Println("Loading...")
		}
		list, err := c.List()
		if err != nil {
			return err
		}

		for _, item := range list {
			fmt.Printf("%10d %s %s\n", item.Size, item.Updated.Format(time.Stamp), item.Path)
		}
	}
	return nil
}

func main() {
	var user, pass string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: %s [upload|delete|sync|push|list] [file or directory]*`, os.Args[0])
		if user != "" {
			fmt.Fprintln(os.Stderr, "warning: username is empty")
		}
		if pass != "" {
			fmt.Fprintln(os.Stderr, "warning: password is empty")
		}
		flag.PrintDefaults()
	}

	flag.StringVar(&user, "user", os.Getenv("NEOCITIES_USER"), "Set username (if not set, it will default to $NEOCITIES_USER)")
	flag.StringVar(&pass, "pass", os.Getenv("NEOCITIES_PASS"), "Set password (if not set, it will default to $NEOCITIES_PASS)")
	flag.BoolVar(&verbose, "v", false, "turn on verbose output")

	flag.Parse()
	if flag.NArg() == 0 || user == "" || pass == "" {
		flag.Usage()
		os.Exit(1)
	}

	err := do(neoc.NewClient(user, pass), strings.ToLower(flag.Arg(0))[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
