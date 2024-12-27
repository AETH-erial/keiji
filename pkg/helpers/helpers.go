package helpers

import (
	"encoding/json"

	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/redis/go-redis/v9"
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
	Ident    string `json:"identifier"`
	Created  string `json:"created"`
	Body     string `json:"body"`
	Category string `json:"category"`
	Sample   string
}

type AdminTables struct {
	Tables []Table `json:"tables"`
}

type Table struct {
	TableName string      `json:"table_name"`
	TableData []TableData `json:"table_data"`
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

	return Document{Ident: ident, Created: ts.String(), Body: body, Category: category}
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
Retrieves the header data from the redis database
*/
func GetHeaders(redisCfg RedisConf) (*HeaderCollection, error) {
	rds := NewRedisClient(redisCfg)
	d, err := rds.Client.Get(rds.ctx, HEADER_KEY).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	header := &HeaderCollection{}
	err = json.Unmarshal([]byte(d), header)
	if err != nil {
		return nil, err
	}
	return header, nil
}

/*
Retrieves the menu elements from the database
*/
func GetMenuLinks(redisCfg RedisConf) (*MenuElement, error) {
	rds := NewRedisClient(redisCfg)
	d, err := rds.Client.Get(rds.ctx, MENU_KEY).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	header := &MenuElement{}
	err = json.Unmarshal([]byte(d), header)
	if err != nil {
		return nil, err
	}
	return header, nil
}

/*
retreives the admin table config from the database
*/
func GetAdminTables(redisCfg RedisConf) (*AdminTables, error) {
	rds := NewRedisClient(redisCfg)
	d, err := rds.Client.Get(rds.ctx, ADMIN_TABLE_KEY).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	tables := &AdminTables{}
	err = json.Unmarshal([]byte(d), tables)
	if err != nil {
		return nil, err
	}
	return tables, nil
}

/*
Place holder func to create the header element in redis
*/
func AddHeaders(h HeaderCollection, redisCfg RedisConf) error {
	rdc := NewRedisClient(redisCfg)
	data, err := json.Marshal(&h)
	if err != nil {
		return err
	}
	err = rdc.Client.Set(rdc.ctx, HEADER_KEY, data, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

/*
Retrieve all documents from the category specified in the argument category

	:param category: the category to get documents from
*/
func GetAllDocuments(category string, redisCfg RedisConf) ([]*Document, error) {
	rdc := NewRedisClient(redisCfg)
	ids, err := rdc.AllDocIds()
	if err != nil {
		return nil, err
	}
	var docs []*Document
	for idx := range ids {
		doc, err := rdc.GetItem(ids[idx])
		if err != nil {
			return nil, err
		}
		if doc.Category != category {
			continue
		}
		docs = append(docs, &Document{
			Ident:   doc.Ident,
			Created: doc.Created,
			Body:    doc.Body,
			Sample:  doc.MakeSample(),
		})
	}
	return docs, nil

}

/*
adds a text post document to the redis database
*/
func AddDocument(d Document, redisCfg RedisConf) error {
	rdc := NewRedisClient(redisCfg)
	return rdc.AddDoc(d)
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
