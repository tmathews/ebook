package ebook

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aymerick/raymond"
)

func GetBaseLen(n int) int {
	s := strconv.Itoa(n)
	return len(s)
}

func GetFiles(dir string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	xs := []os.FileInfo{}
	for _, x := range files {
		if x.IsDir() {
			continue
		}
		name := x.Name()
		ext := filepath.Ext(name)
		if ext != ".jpeg" && ext != ".png" {
			continue
		}
		xs = append(xs, x)
	}
	return xs, nil
}

func ZipTemplate(a *zip.Writer, tpl *raymond.Template, scope interface{}, f string) error {
	str, err := tpl.Exec(scope)
	if err != nil {
		return err
	}
	return ZipString(a, f, str)
}

func ZipString(a *zip.Writer, f string, str string) error {
	header := zip.FileHeader{
		Name:     f,
		Method:   zip.Deflate,
		Modified: time.Now(),
	}
	out, err := a.CreateHeader(&header)
	if err != nil {
		return err
	}
	_, err = io.WriteString(out, str)
	if err != nil {
		return err
	}
	return nil
}

func ZipFile(archive *zip.Writer, src string, dst string) error {
	header := zip.FileHeader{
		Name:     dst,
		Method:   zip.Deflate,
		Modified: time.Now(),
	}
	out, err := archive.CreateHeader(&header)
	if err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}
