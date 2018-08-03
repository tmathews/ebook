package ebook

import (
	"encoding/json"
	"mime"
	"os"
	"path/filepath"
	"time"

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
