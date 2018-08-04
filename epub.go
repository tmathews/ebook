package ebook

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
)

var bintxt = map[string]string{
	"container.xml": `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
<rootfiles>
<rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
</rootfiles>
</container>`,
	"style.css": `@page {
	margin: 0;
}

body {
	display: block;
	margin: 0;
	padding: 0;
}`,
	"template.xhtml": `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
	<head>
		<title>{{.Number}}</title>
	</head>
	<body>
		<div>
			<img src="../Images/{{.Filename}}"/>
		</div>
	</body>
</html>`,
	"content.opf": `<?xml version="1.0" encoding="UTF-8"?>
<package version="3.0" unique-identifier="BookID" xmlns="http://www.idpf.org/2007/opf">
	<metadata xmlns:opf="http://www.idpf.org/2007/opf" xmlns:dc="http://purl.org/dc/elements/1.1/">
		<dc:title>{{.Title}}</dc:title>
		<dc:description>{{.Description}}</dc:description>
		<dc:language>{{.Locale}}</dc:language>
		<dc:identifier id="BookID">urn:uuid:{{.Id}}</dc:identifier>
		<dc:contributor id="contributor">{{.Contributor}}</dc:contributor>
		<dc:creator>{{.Creator}}</dc:creator>
		{{range .Subjects}}
		<dc:subject>{{.}}</dc:subject>
		{{end}}
		<meta property="dcterms:modified">{{.DateModifiedStr}}</meta>
		<meta name="cover" content="cover"/>
		<meta property="rendition:orientation">portrait</meta>
		<meta property="rendition:spread">portrait</meta>
		<meta property="rendition:layout">pre-paginated</meta>
		<meta name="original-resolution" content="1072x1448"/>
		<meta name="book-type" content="{{.Type}}"/>
		<meta name="RegionMagnification" content="true"/>
		<meta name="primary-writing-mode" content="{{.WritingMode}}"/>
		<meta name="zero-gutter" content="{{.ZeroGutter}}"/>
		<meta name="zero-margin" content="{{.ZeroMargin}}"/>
		<meta name="ke-border-color" content="#ffffff"/>
		<meta name="ke-border-width" content="0"/>
	</metadata>
	<manifest>
		<item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
		<item id="nav" href="nav.xhtml" properties="nav" media-type="application/xhtml+xml"/>
		<item id="cover" href="Images/{{.Cover.Filename}}" media-type="{{.Cover.MimeType}}" properties="cover-image"/>
		{{range .Pages}}
		<item id="page_Images_{{.Number}}" href="Text/{{.Number}}.xhtml" media-type="application/xhtml+xml"/>
		<item id="img_Images_{{.Number}}" href="Images/{{.Filename}}" media-type="{{.MimeType}}"/>
		{{end}}
		<item id="css" href="Text/style.css" media-type="text/css"/>
	</manifest>
	<spine page-progression-direction="{{.Direction}}" toc="ncx">
		{{range .Pages}}
		<itemref idref="page_Images_{{.Number}}"/>
		{{end}}
	</spine>
</package>`,
	"nav.xhtml": `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
	<head>
		<title>{{.Title}}</title>
		<meta charset="utf-8"/>
	</head>
	<body>
		<nav xmlns:epub="http://www.idpf.org/2007/ops" epub:type="toc" id="toc">
			<ol>
				<li><a href="Text/{{.FirstPage.Number}}.xhtml">{{.Title}}</a></li>
			</ol>
		</nav>
		<nav epub:type="page-list">
			<ol>
				<li><a href="Text/{{.FirstPage.Number}}.xhtml">{{.Title}}</a></li>
			</ol>
		</nav>
	</body>
</html>`,
	"toc.ncx": `<?xml version="1.0" encoding="UTF-8"?>
<ncx version="2005-1" xml:lang="en-US" xmlns="http://www.daisy.org/z3986/2005/ncx/">
	<head>
		<meta name="dtb:uid" content="urn:uuid:{{.Id}}"/>
		<meta name="dtb:depth" content="1"/>
		<meta name="dtb:totalPageCount" content="0"/>
		<meta name="dtb:maxPageNumber" content="0"/>
		<meta name="generated" content="true"/>
	</head>
	<docTitle><text>{{.Title}}</text></docTitle>
	<navMap>
		<navPoint id="Text">
			<navLabel>
				<text>{{.Title}}</text>
			</navLabel>
			<content src="Text/{{.FirstPage.Number}}.xhtml"/>
		</navPoint>
	</navMap>
</ncx>`,
}

func getTemplates() (*template.Template, error) {
	tpls := template.New("global")
	for name, txt := range bintxt {
		tpl := tpls.New(name)
		_, err := tpl.Parse(txt)
		if err != nil {
			return nil, err
		}
	}
	return tpls, nil
}

func EPubFromDir(book Book, dir string) (tmpf string, err error) {
	tpls, err := getTemplates()
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
	err = ZipString(archive, "mimetype", "application/epub+zip")
	if err != nil {
		return
	}
	err = ZipString(archive, "META-INF/container.xml", bintxt["container.xml"])
	if err != nil {
		return
	}
	err = ZipString(archive, "OEBPS/Text/style.css", bintxt["style.css"])
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
		err = ZipTemplate(archive, tpls.Lookup("template.xhtml"), page,
			filepath.Join("OEBPS", "Text", fmt.Sprintf(fmtPageFile, num, ".xhtml")))
		if err != nil {
			return
		}
	}
	book.Pages = pages

	err = ZipTemplate(archive, tpls.Lookup("content.opf"), book,
		filepath.Join("OEBPS/content.opf"))
	if err != nil {
		return
	}
	err = ZipTemplate(archive, tpls.Lookup("nav.xhtml"), book,
		filepath.Join("OEBPS/nav.xhtml"))
	if err != nil {
		return
	}
	err = ZipTemplate(archive, tpls.Lookup("toc.ncx"), book,
		filepath.Join("OEBPS/toc.ncx"))
	if err != nil {
		return
	}
	err = archive.Close()
	if err != nil {
		return
	}
	return
}
