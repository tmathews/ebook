package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tmathews/ebook-gen"
)

func main() {
	var dir string
	var out string
	var meta string

	flag.StringVar(&dir, "dir", "", "Source directory to compile.")
	flag.StringVar(&out, "o", "", "Desired filepath with format. (EPUB, MOBI, CBZ) (MOBI requires kindlegen in your $PATH)")
	flag.StringVar(&meta, "meta", "", "Apply metadata from a file. (JSON supported)")
	flag.Parse()

	book := ebook.NewBook()
	book.Title = filepath.Base(dir)

	if meta != "" {
		err := book.FromJSON(meta)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	ext := filepath.Ext(out)
	if ext == ".epub" || ext == ".mobi" {
		GenEpub(book, dir, out, ext == ".mobi")
	} else if ext == ".cbz" {
		GenCbz(book, dir, out)
	} else {
		fmt.Println("Nothing to do.")
	}
}

func GenCbz(book ebook.Book, dir string, out string) {
	tmpf, err := ebook.CbzFromDir(book, dir)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = os.Rename(tmpf, out)
	if err != nil {
		fmt.Println(err)
	}
}

func GenEpub(book ebook.Book, dir string, out string, kindle bool) {
	tmpf, err := ebook.EPubFromDir(book, dir)
	if err != nil {
		fmt.Println(err)
		return
	}
	if kindle {
		err = ToKindle(tmpf, out)
		os.Remove(tmpf)
	} else {
		err = os.Rename(tmpf, out)
	}
	if err != nil {
		fmt.Println(err)
	}
}

func ToKindle(in string, out string) error {
	cmd := exec.Command("kindlegen", in)
	cmd.Start()
	err := cmd.Wait()
	if err != nil {
		return err
	}
	return os.Rename(in, out)
}