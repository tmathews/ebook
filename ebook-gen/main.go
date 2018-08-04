package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tmathews/ebook"
)

const Usage = `ebook-gen is a tool to convert an image directory into an ebook.

Usage:

	ebook-gen [-m] directory [output]

"directory" is the source of files you wish to convert.

"output" is the path where you want the desired book to be placed. If nothing is
provided it will be assumed to use the current location.

"-f" Specify which format you wish to use (EPUB, CBZ, MOBI). If not provided, it
will try to auto detect from the output path. Generating MOBI files requires you
to have "kindlegen" in your $PATH.

"-m" Allows you specify a JSON file to load metadata from.
`

const (
	ErrArg  = 1
	ErrMeta = 2
	ErrSys  = 3
)

func main() {
	os.Exit(Do())
}

func Do() int {
	var meta string
	var format string

	flag.Usage = func() {
		fmt.Println(Usage)
	}
	flag.StringVar(&meta, "m", "", "Apply metadata from a file. (JSON supported)")
	flag.StringVar(&format, "f", "", "Specify the output format. (EPUB, CBZ, MOBI)")
	flag.Parse()

	dir := flag.Arg(0)
	out := flag.Arg(1)
	if dir == "" {
		flag.Usage()
		return ErrArg
	}

	// Ensure the source is a directory
	fi, err := os.Stat(dir)
	if err != nil {
		log.Fatal(err)
		return ErrSys
	}
	if !fi.Mode().IsDir() {
		log.Fatal(errors.New("Provided directory path is not a directory."))
		return ErrSys
	}

	// Clean the provided directory path for usage.
	dir, err = filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
		return ErrArg
	}

	format = ParseFormat(format)
	dirname := filepath.Base(dir)

	// Auto create filename based on directory name
	if out == "" {
		if format == "" {
			log.Fatal(errors.New("No format provided."))
			return ErrArg
		}
		out = filepath.Join(".", dirname+"."+format)
	} else { // Auto assume format by input
		desiredFormat := ParseFormat(filepath.Ext(out))
		if format == "" && desiredFormat == "" {
			log.Fatal(errors.New("No format provided for output location."))
			return ErrArg
		} else if format == "" && desiredFormat != "" {
			format = desiredFormat
		}
	}

	// Auto assume filepath if provided is a directory
	// Do this here because of OOO (out path conflicts)
	fi, err = os.Stat(out)
	if err != nil {
		log.Fatal(err)
		return ErrSys
	}
	if fi.Mode().IsDir() {
		out = filepath.Join(out, dirname+"."+format)
	}

	book := ebook.NewBook()
	book.Title = dirname

	if meta != "" {
		err := book.FromJSON(meta)
		if err != nil {
			log.Fatal(err)
			return ErrMeta
		}
	}

	if format == "epub" || format == "mobi" {
		return GenEpub(book, dir, out, format == "mobi")
	} else if format == "cbz" {
		return GenCbz(book, dir, out)
	} else {
		flag.Usage()
		return ErrArg
	}
	return 0
}

func ParseFormat(str string) string {
	str = strings.Replace(strings.ToLower(str), ".", "", -1)
	if str != "epub" && str != "cbz" && str != "mobi" {
		return ""
	}
	return str
}

func GenCbz(book ebook.Book, dir string, out string) int {
	tmpf, err := ebook.CbzFromDir(book, dir)
	if err != nil {
		log.Fatal(err)
		return ErrSys
	}
	err = os.Rename(tmpf, out)
	if err != nil {
		log.Fatal(err)
		return ErrSys
	}
	return 0
}

func GenEpub(book ebook.Book, dir string, out string, kindle bool) int {
	tmpf, err := ebook.EPubFromDir(book, dir)
	if err != nil {
		log.Fatal(err)
		return ErrSys
	}
	if kindle {
		err = ToKindle(tmpf, out)
		os.Remove(tmpf)
	} else {
		err = os.Rename(tmpf, out)
	}
	if err != nil {
		log.Fatal(err)
		return ErrSys
	}
	return 0
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
