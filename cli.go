package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	user := flag.String("user", os.Getenv("NCU"),
		"Set username (if not set, it will default to $NCU)")
	pass := flag.String("pass", os.Getenv("NCP"),
		"Set password (if not set, it will default to $NCP)")
	verb := flag.Bool("v", false, "output verbosely")

	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Printf("usage: %s [upload|delete|sync|push|list] [file o. dir]\n", os.Args[0])
		fmt.Printf("       Use -h for information on username and passwords\n")
		os.Exit(1)
	}

	if *user == "" {
		fmt.Printf("Username not defined\n")
		os.Exit(1)
	}

	if *pass == "" {
		fmt.Printf("Password not defined\n")
		os.Exit(1)
	}

	var (
		err  error
		size int
		list []Item
	)

	switch strings.ToLower(flag.Arg(0))[0] {
	case 'u': // upload
		if flag.NArg() < 2 {
			fmt.Printf("UPLOAD request requires arguments\n")
			os.Exit(3)
		}

		if *verb {
			fmt.Println("Uploading...")
		}
		err = Upload(flag.Args()[1:], *user, *pass)
	case 'd': // delete
		if flag.NArg() < 2 {
			fmt.Printf("DELETE request requires arguments\n")
			os.Exit(3)
		}

		if *verb {
			fmt.Println("Deleting...")
		}
		err = Delete(flag.Args()[1:], *user, *pass)
	case 's': // sync
		ch := make(chan string)
		size, err = Sync(flag.Arg(1), ch, *user, *pass)

		if *verb {
			fmt.Println("Downloading...")
		}

		for i := 1; i <= size; i++ {
			msg, ok := <-ch

			if !ok {
				break
			}

			if msg == "" {
				continue
			}

			if *verb {
				fmt.Printf("[%d/%d] %s...\n", i, size, msg)
			}
		}
	case 'p': // push
		ch := make(chan string, 1<<5)
		err = Push(flag.Arg(1), ch, *user, *pass)

		if *verb {
			fmt.Println("Pushing...")
		}

		var flist []string
		for {
			msg, ok := <-ch
			if !ok {
				break
			}

			flist = append(flist, msg)
			if *verb {
				fmt.Println(msg)
			}
		}

		Upload(flist, *user, *pass)
	case 'l': // list
		if *verb {
			fmt.Println("Loading...")
		}
		list, err = List(*user, *pass)

		for _, itm := range list {
			fmt.Printf("%10d %s %s\n", itm.Size, itm.Updated.Format(time.Stamp), itm.Path)
		}
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if *verb {
		fmt.Println("Done")
	}

}
