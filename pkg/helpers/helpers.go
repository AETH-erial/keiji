package helpers

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

const TECHNICAL = "technical"
const CONFIGURATION = "configuration"
const BLOG = "blog"
const CREATIVE = "creative"
const DIGITAL_ART = "digital_art"
const HOMEPAGE = "homepage"

var Topics = []string{
	TECHNICAL,
	BLOG,
	CREATIVE,
	HOMEPAGE,
}

type HeaderCollection struct {
	Category string       `json:"category"`
	Elements []HeaderElem `json:"elements"`
}

type HeaderElem struct {
	Png  string `json:"png"`
	Link string `json:"link"`
}

type ImageElement struct {
	ImgUrl string `json:"img_url"`
}

type MenuElement struct {
	Png       string         `json:"png"`
	Category  string         `json:"category"`
	MenuLinks []MenuLinkPair `json:"menu_links"`
}

type DocumentOld struct {
	Ident    Identifier `json:"identifier"`
	Created  string     `json:"created"`
	Body     string     `json:"body"`
	Category string     `json:"category"`
	Sample   string
}

type AdminPage struct {
	Tables map[string][]TableData `json:"tables"`
}

type TableData struct { // TODO: add this to the database io interface
	DisplayName string `json:"display_name"`
	Link        string `json:"link"`
}

/*
	 convert markdown to html
		:param md: the byte array containing the Markdown to convert
*/
func MdToHTML(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}
