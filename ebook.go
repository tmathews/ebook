package ebook

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aymerick/raymond"
	"github.com/google/uuid"
)

type Direction string

const (
	DirectionLeft  Direction = "ltr"
	DirectionRight Direction = "rtl"
)

type WritingMode string

const (
	WritingModeHLR WritingMode = "horizontal-lr"
	WritingModeHRL WritingMode = "horizontal-rl"
)

type Orientation string

const (
	OrientationLandscape Orientation = "landscape"
	OrientationPortrait  Orientation = "portrait"
)

type BookType string

const (
	TypeComic BookType = "comic"
)

type Locale string

const (
	LocaleEnglish    Locale = "en"
	LocaleGerman     Locale = "de"
	LocaleFrench     Locale = "fr"
	LocaleItalian    Locale = "it"
	LocaleSpanish    Locale = "es"
	LocaleChinese    Locale = "zh"
	LocaleJapanese   Locale = "ja"
	LocalePortuguese Locale = "pt"
	LocaleRussian    Locale = "ru"
	LocaleDutch      Locale = "nl"
)

type Book struct {
	Contributor  string
	Cover        Page
	Creator      string
	DateModified time.Time
	Direction    Direction
	Id           string
	Layout       string
	Locale       Locale
	Orientation  Orientation
	Pages        []Page
	Spread       string
	Subjects     []string
	Title        string
	Type         BookType
	WritingMode  WritingMode
	ZeroGutter   bool
	ZeroMargin   bool
}

func (b Book) DateModifiedStr() string {
	return b.DateModified.Format(time.RFC3339)
}

func (b Book) FirstPage() Page {
	if len(b.Pages) < 1 {
		return Page{}
	}
	return b.Pages[0]
}

func (b Book) Language() Locale {
	return b.Locale
}

func (b *Book) FromJSON(f string) (err error) {
	file, err := os.Open(f)
	if err != nil {
		return
	}
	err = json.NewDecoder(file).Decode(&b)
	if err != nil {
		return
	}
	return nil
}

type Page struct {
	Number   string
	Filename string
}

func (p Page) MimeType() string {
	return mime.TypeByExtension(filepath.Ext(p.Filename))
}

func NewBook() Book {
	return Book{
		Locale:       LocaleEnglish,
		DateModified: time.Now(),
		Direction:    DirectionLeft,
		Id:           uuid.New().String(),
		Orientation:  OrientationPortrait,
		Type:         TypeComic,
		WritingMode:  WritingModeHRL,
		ZeroGutter:   true,
		ZeroMargin:   true,
	}
}

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

func RenderTemplate(tpl *raymond.Template, scope interface{}, filename string) error {
	str, err := tpl.Exec(scope)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, []byte(str), 664)
	if err != nil {
		return err
	}
	return nil
}

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
