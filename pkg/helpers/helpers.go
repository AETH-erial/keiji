package helpers

import (

	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

const HEADER_KEY = "header-links"
const MENU_KEY = "menu-config"
const ADMIN_TABLE_KEY = "admin-tables"

const TECHNICAL = "technical"
const CONFIGURATION = "configuration"
const BLOG = "blog"
const CREATIVE = "creative"
const DIGITAL_ART = "digital_art"

var Topics = []string{
	TECHNICAL,
	BLOG,
	CREATIVE,
}

var TopicMap = map[string]string{
	TECHNICAL: TECHNICAL,
	BLOG:      BLOG,
	CREATIVE:  CREATIVE,
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
	Ident    Identifier`json:"identifier"`
	Created  string `json:"created"`
	Body     string `json:"body"`
	Category string `json:"category"`
	Sample   string
}

type AdminPage struct {
	Tables map[string][]TableData `json:"tables"`
}


type TableData struct { // TODO: add this to the database io interface 
	DisplayName string `json:"display_name"`
	Link        string `json:"link"`
}

func NewDocument(ident string, created *time.Time, body string, category string) Document {

	var ts time.Time
	if created == nil {
		rn := time.Now()
		ts = time.Date(rn.Year(), rn.Month(), rn.Day(), rn.Hour(), rn.Minute(),
			rn.Second(), rn.Nanosecond(), rn.Location())
	} else {
		ts = *created
	}

	return Document{Ident: Identifier(ident), Created: ts.String(), Body: body, Category: category}
}

type DocumentUpload struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Text     string `json:"text"`
}

type HeaderIo interface {
	GetHeaders() (*HeaderCollection, error)
	AddHeaders(HeaderCollection) error
	GetMenuLinks() (*MenuElement, error)
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
