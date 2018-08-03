package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tmathews/ebook-gen"
)

func main() {
	var dir string
	var out string
	var kindle bool
	var meta string

	flag.StringVar(&dir, "dir", "", "Source directory to compile.")
	flag.StringVar(&out, "o", "", "Desired filename without the extension.")
	flag.StringVar(&meta, "meta", "", "Apply metadata from a file. (JSON supported)")
	flag.BoolVar(&kindle, "kindle", false, "Generate to Kindle MOBI format. (Requires kindlegen in your $PATH)")
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

	tmpf, err := ebook.EPubFromDir(book, dir)
	if err != nil {
		fmt.Println(err)
		return
	}
	if out == "" {
		out = filepath.Join(".", book.Id)
	}
	if kindle {
		err = ToKindle(tmpf, out + ".mobi")
		os.Remove(tmpf)
	} else {
		err = os.Rename(tmpf, out + ".epub")
	}
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ToKindle(in string, out string) error {
	cmd := exec.Command("kindlegen", in)
	cmd.Start()
	err := cmd.Wait()
	if err != nil {
		return err
	}

	dir := filepath.Dir(in)
	filename := filepath.Base(in)
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	return os.Rename(filepath.Join(dir, name + ".mobi"), out)
}