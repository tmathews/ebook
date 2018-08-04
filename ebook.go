package ebook

import (
	"encoding/json"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type M map[string]interface{}

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
	LocaleChinese    Locale = "zh"
	LocaleDutch      Locale = "nl"
	LocaleEnglish    Locale = "en"
	LocaleFrench     Locale = "fr"
	LocaleGerman     Locale = "de"
	LocaleItalian    Locale = "it"
	LocaleJapanese   Locale = "ja"
	LocalePortuguese Locale = "pt"
	LocaleRussian    Locale = "ru"
	LocaleSpanish    Locale = "es"
)

var LocaleLanguageMap = map[Locale]string{
	LocaleChinese:    "Chinese",
	LocaleDutch:      "Dutch",
	LocaleEnglish:    "English",
	LocaleFrench:     "French",
	LocaleGerman:     "German",
	LocaleItalian:    "Italian",
	LocaleJapanese:   "Japanese",
	LocalePortuguese: "Portuguese",
	LocaleRussian:    "Russian",
	LocaleSpanish:    "Spanish",
}

type Book struct {
	Comments        string
	Contributor     string
	Country         string
	Cover           Page
	Creator         string
	Credits         []Credit
	DateModified    time.Time
	Direction       Direction
	Description     string
	Genre           string
	Id              string
	IssueCount      int
	IssueNumber     int
	Layout          string
	Locale          Locale
	Orientation     Orientation
	Pages           []Page
	PublicationDate time.Time
	Publisher       string
	Rating          int
	Series          string
	Spread          string
	Subjects        []string
	Tags            []string
	Title           string
	Type            BookType
	VolumeCount     int
	VolumeNumber    int
	WritingMode     WritingMode
	ZeroGutter      bool
	ZeroMargin      bool
}

func (b Book) FirstPage() Page {
	if len(b.Pages) < 1 {
		return Page{}
	}
	return b.Pages[0]
}

func (b Book) DateModifiedStr() string {
	return b.DateModified.Format(time.RFC3339)
}

func (b Book) LanguageStr() string {
	return LocaleLanguageMap[b.Locale]
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

type Credit struct {
	IsPrimary bool
	Name      string
	Role      string
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
		DateModified: time.Now(),
		Direction:    DirectionLeft,
		Id:           uuid.New().String(),
		Locale:       LocaleEnglish,
		Orientation:  OrientationPortrait,
		Type:         TypeComic,
		WritingMode:  WritingModeHRL,
		ZeroGutter:   true,
		ZeroMargin:   true,
	}
}
