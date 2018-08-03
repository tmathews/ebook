package ebook

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aymerick/raymond"
)

func EPubFromDir(book Book, dir string) (tmpf string, err error) {
	tplPage, err := raymond.ParseFile("./template/OEBPS/Text/template.xhtml")
	if err != nil {
		return
	}
	tplCont, err := raymond.ParseFile("./template/OEBPS/content.opf")
	if err != nil {
		return
	}
	tplNav, err := raymond.ParseFile("./template/OEBPS/nav.xhtml")
	if err != nil {
		return
	}
	tplToc, err := raymond.ParseFile("./template/OEBPS/toc.ncx")
	if err != nil {
		return
	}

	files, err := GetFiles(dir)
	if err != nil {
		return
	}

	tmpf = filepath.Join(os.TempDir(), book.Id+".epub")
	file, err := os.Create(tmpf)
	if err != nil {
		return
	}
	defer file.Close()

	archive := zip.NewWriter(file)
	err = ZipFile(archive, "./template/mimetype", "mimetype")
	if err != nil {
		return
	}
	err = ZipFile(archive, "./template/META-INF/container.xml", "META-INF/container.xml")
	if err != nil {
		return
	}
	err = ZipFile(archive, "./template/OEBPS/Text/style.css", "OEBPS/Text/style.css")
	if err != nil {
		return
	}

	length := len(files)
	fmtPage := "%0" + strconv.Itoa(GetBaseLen(length)) + "d"
	fmtPageFile := fmtPage + "%v"
	pages := []Page{}
	for i, x := range files {
		name := x.Name()
		ext := filepath.Ext(name)
		num := i + 1
		page := Page{
			Number:   fmt.Sprintf(fmtPage, num),
			Filename: fmt.Sprintf(fmtPageFile, num, ext),
		}
		imgpath := filepath.Join(dir, name)

		err = ZipFile(archive, imgpath, filepath.Join("OEBPS", "Images", page.Filename))
		if err != nil {
			return
		}
		if i == 0 {
			book.Cover = Page{
				Number:   page.Number,
				Filename: "cover" + ext,
			}
			err = ZipFile(archive, imgpath, filepath.Join("OEBPS", "Images", book.Cover.Filename))
			if err != nil {
				return
			}
		}
		pages = append(pages, page)
		err = ZipTemplate(archive, tplPage, page, filepath.Join("OEBPS", "Text",
			fmt.Sprintf(fmtPageFile, num, ".xhtml")))
		if err != nil {
			return
		}
	}
	book.Pages = pages

	err = ZipTemplate(archive, tplCont, book, filepath.Join("OEBPS/content.opf"))
	if err != nil {
		return
	}
	err = ZipTemplate(archive, tplNav, book, filepath.Join("OEBPS/nav.xhtml"))
	if err != nil {
		return
	}
	err = ZipTemplate(archive, tplToc, book, filepath.Join("OEBPS/toc.ncx"))
	if err != nil {
		return
	}
	err = archive.Close()
	if err != nil {
		return
	}
	return
}
