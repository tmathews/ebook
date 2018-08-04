package ebook

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func GetBookCbzMetadata(book Book) map[string]interface{} {
	credits := []M{}
	for _, c := range book.Credits {
		credits = append(credits, M{
			"person":  c.Name,
			"primary": c.IsPrimary,
			"role":    c.Role,
		})
	}
	return M{
		"appId":        "ebook-gen",
		"lastModified": book.DateModifiedStr(),
		"ComicBookInfo/1.0": M{
			"comments":         book.Comments,
			"country":          book.Country,
			"credits":          credits,
			"genre":            book.Genre,
			"issue":            book.IssueNumber,
			"language":         book.LanguageStr(),
			"numberOfIsseus":   book.IssueCount,
			"numberOfVolumes":  book.VolumeCount,
			"publicationMonth": book.PublicationDate.Month(),
			"publicationYear":  book.PublicationDate.Year(),
			"publisher":        book.Publisher,
			"rating":           book.Rating,
			"series":           book.Series,
			"tags":             book.Tags,
			"title":            book.Title,
			"volume":           book.VolumeNumber,
		},
	}
}

func CbzFromDir(book Book, dir string) (tmpf string, err error) {
	files, err := GetFiles(dir)
	if err != nil {
		return
	}
	tmpf = filepath.Join(os.TempDir(), book.Id+".cbz")
	file, err := os.Create(tmpf)
	if err != nil {
		return
	}
	defer file.Close()
	archive := zip.NewWriter(file)

	meta, err := json.Marshal(GetBookCbzMetadata(book))
	if err != nil {
		return
	}
	err = archive.SetComment(string(meta))
	if err != nil {
		return
	}

	length := len(files)
	fmtPage := "%0" + strconv.Itoa(GetBaseLen(length)) + "d"
	fmtPageFile := fmtPage + "%v"
	for i, x := range files {
		name := x.Name()
		ext := filepath.Ext(name)
		num := i + 1
		page := Page{
			Number:   fmt.Sprintf(fmtPage, num),
			Filename: fmt.Sprintf(fmtPageFile, num, ext),
		}
		imgpath := filepath.Join(dir, name)

		err = ZipFile(archive, imgpath, page.Filename)
		if err != nil {
			return
		}
	}

	err = archive.Close()
	if err != nil {
		return
	}
	return
}
