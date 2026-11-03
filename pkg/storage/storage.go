package storage

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"time"

	"git.aetherial.dev/aeth/keiji/pkg/env"
	"github.com/google/uuid"
)

const TECHNICAL = "technical"
const CONFIGURATION = "configuration"
const BLOG = "blog"
const CREATIVE = "creative"
const DIGITAL_ART = "digital_art"
const MISC_ASSET = "misc_asset"
const HOMEPAGE = "homepage"

var Topics = []string{
	TECHNICAL,
	BLOG,
	CREATIVE,
	HOMEPAGE,
}

var ImageCategories = []string{
	MISC_ASSET,
	DIGITAL_ART,
}

type DatabaseSchema struct {
	// Gotta figure out what this looks like
	// so that the ExtractAll() function gets
	// all of the data from the database

}

type MenuElement struct {
	Png       string     `json:"png"`
	Category  string     `json:"category"`
	MenuLinks []LinkPair `json:"menu_links"`
}
type AdminPage struct {
	Tables map[string][]TableData `json:"tables"`
}

type TableData struct { // TODO: add this to the database io interface
	DisplayName string `json:"display_name"`
	Link        string `json:"link"`
}

type LinkPair struct {
	Link string `json:"link"`
	Text string `json:"text"`
}

type NavBarItem struct {
	Png      []byte                `json:"png"`
	File     *multipart.FileHeader `form:"file"`
	Link     string                `json:"link" form:"link"`
	Redirect string                `json:"redirect" form:"redirect"`
}

type Asset struct {
	Name string
	Data []byte
}

type Identifier string

type Document struct {
	Row      int
	Ident    Identifier `json:"id"`
	Title    string     `json:"title"`
	Created  string     `json:"created"`
	Body     string     `json:"body"`
	Category string     `json:"category"`
	Sample   string     `json:"sample"`
}

/*
Truncates a text post into a 256 character long 'sample' for displaying posts
*/
func (d *Document) MakeSample() string {
	t := strings.Split(d.Body, "")
	var sample []string
	if len(d.Body) < 256 {
		return d.Body
	}
	for i := 0; i < 256; i++ {
		sample = append(sample, t[i])
	}
	sample = append(sample, " ...")
	return strings.Join(sample, "")

}

type Image struct {
	Ident     Identifier `json:"identifier"`
	Location  string
	Title     string                `json:"title" form:"title"`
	File      *multipart.FileHeader `form:"file"`
	Desc      string                `json:"description" form:"description"`
	Created   string
	Category  string `json:"category" form:"category"`
	MediaType string `json:"media_type"`
	Data      []byte
}

type DocumentIO interface {
	GetDocument(id Identifier) (Document, error)
	GetImage(id Identifier) (Image, error)
	GetAllImages() []Image
	GetImagesByCategory(category string) []Image
	UpdateDocument(doc Document) error
	DeleteDocument(id Identifier) error
	UpdateDigitalArt(id Identifier, title, desc string) error
	ArchiveDigitalArt(id Identifier) error
	DeleteNavbarItem(id Identifier) error
	AddDocument(doc Document) (Identifier, error)
	AddImage(data []byte, title, desc, category, mediaType string) (Identifier, error)
	AddAsset(name string, data []byte) error
	AddAdminTableEntry(TableData, string) error
	AddNavbarItem(NavBarItem) error
	AddMenuItem(LinkPair) error
	GetByCategory(category string) []Document
	AllDocuments() []Document
	GetDropdownElements() []LinkPair
	GetNavBarLinks() []NavBarItem
	GetAssets() []Asset
	GetAdminTables() AdminPage
}

// DocumentListReader provides list operations that distinguish database failures
// from empty results. Request handlers should prefer these methods.
type DocumentListReader interface {
	GetDropdownElementsWithError() ([]LinkPair, error)
	GetNavBarLinksWithError() ([]NavBarItem, error)
	GetAssetsWithError() ([]Asset, error)
	GetAdminTablesWithError() (AdminPage, error)
	GetByCategoryWithError(string) ([]Document, error)
	AllDocumentsWithError() ([]Document, error)
	GetAllImagesWithError() ([]Image, error)
	GetImagesByCategoryWithError(string) ([]Image, error)
}

var (
	ErrDuplicate    = errors.New("record already exists")
	ErrNotExists    = errors.New("row not exists")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)

type SQLiteRepo struct {
	db      *sql.DB
	imageIO ImageIO
}

type ImageIO interface {
	Put([]byte, Identifier) error
	Get(Identifier) ([]byte, error)
}

// ReadSeekCloser is a seekable media stream that must be closed by its caller.
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// ImageOpener is an optional streaming extension to ImageIO.
type ImageOpener interface {
	Open(Identifier) (ReadSeekCloser, error)
}

type FilesystemImageIO struct {
	RootDir string
}

/*
create a new FilesystemImageIO struct
*/
func MustNewFilesystemImageIO(path string) FilesystemImageIO {
	var filemode os.FileMode
	filemode = 0775
	err := os.MkdirAll(path, filemode)
	if err != nil {
		log.Fatalf("Error creating path: '%s' with permissions: '%d'. Error: %s\n", path, filemode, err.Error())
	}
	return FilesystemImageIO{RootDir: path}
}

/*
Put a data blob on the filesystem

	:param b: the
*/
func (f FilesystemImageIO) Put(b []byte, id Identifier) error {
	if err := validateImageIdentifier(id); err != nil {
		return err
	}
	fh, err := os.OpenFile(path.Join(f.RootDir, string(id)), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fh.Close()
	_, err = fh.Write(b)
	if err != nil {
		return err
	}
	return nil
}

/*
Get a data blob from the filesystem

	:param id: the identifier of the image to retrieve
*/
func (f FilesystemImageIO) Get(id Identifier) ([]byte, error) {
	fh, err := f.Open(id)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	b, err := io.ReadAll(fh)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Open opens a filesystem image without reading it into memory.
func (f FilesystemImageIO) Open(id Identifier) (ReadSeekCloser, error) {
	if err := validateImageIdentifier(id); err != nil {
		return nil, err
	}
	return os.Open(path.Join(f.RootDir, string(id)))
}

func validateImageIdentifier(id Identifier) error {
	name := string(id)
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00") {
		return fmt.Errorf("invalid image identifier %q", name)
	}
	return nil
}

// Instantiate a new SQLiteRepo struct
func NewSQLiteRepo(db *sql.DB, imgIo ImageIO) *SQLiteRepo {
	return &SQLiteRepo{
		db:      db,
		imageIO: imgIo,
	}

}

// Creates a new SQL table for text posts
func (r *SQLiteRepo) Migrate(seedQueries []string) error {
	for i := range seedQueries {
		_, err := r.db.Exec(seedQueries[i])
		if err != nil {
			return err
		}
	}
	// CREATE TABLE IF NOT EXISTS does not update an existing images table.
	// Check by name so migration remains safe to run repeatedly.
	rows, err := r.db.Query("PRAGMA table_info(images)")
	if err != nil {
		return err
	}
	var hasImages, hasMediaType bool
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			rows.Close()
			return err
		}
		hasImages = true
		if name == "media_type" {
			hasMediaType = true
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if hasImages && !hasMediaType {
		_, err = r.db.Exec("ALTER TABLE images ADD COLUMN media_type TEXT NOT NULL DEFAULT ''")
		return err
	}
	return nil
}

/*
Get all dropdown menu elements. Returns a list of LinkPair structs with the text and redirect location
*/
func (s *SQLiteRepo) GetDropdownElements() []LinkPair {
	items, err := s.GetDropdownElementsWithError()
	if err != nil {
		log.Printf("get dropdown elements: %v", err)
	}
	return items
}

func (s *SQLiteRepo) GetDropdownElementsWithError() ([]LinkPair, error) {
	rows, err := s.db.Query("SELECT * FROM menu")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LinkPair
	for rows.Next() {
		var id int
		var item LinkPair
		if err := rows.Scan(&id, &item.Link, &item.Text); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

/*
Retrieve a dropdown element by its text name on the UI
*/
func (s *SQLiteRepo) GetMenuItemByName(link, text string) (LinkPair, bool) {
	fmt.Printf("From GetMenuItemByName: Link: %s Text: %s\n", link, text)
	rows := s.db.QueryRow("SELECT * FROM menu WHERE link = ? AND text = ?", link, text)
	var item LinkPair
	var id int
	err := rows.Scan(&id, &item.Link, &item.Text)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return item, false
		} else {
			return item, false
		}
	}
	return item, true

}

// get Admin table entry by its display name, link, and category.
func (s *SQLiteRepo) GetAdminTableEntry(displayName, link, category string) (TableData, bool) {
	rows := s.db.QueryRow("SELECT * FROM admin WHERE display_name = ? AND link = ? AND category = ?", displayName, link, category)
	var item TableData
	var id int
	err := rows.Scan(&id, &item.DisplayName, &item.Link, &category)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return item, false
		}
		log.Printf("Error getting admin table entry: %s", err.Error())
		return item, false
	}
	return item, true

}

// get navbar entry.
func (s *SQLiteRepo) GetNavbarLink(link, redirect string) (NavBarItem, bool) {
	rows := s.db.QueryRow("SELECT * FROM navbar WHERE link = ? AND redirect = ?", link, redirect)
	var item NavBarItem
	var id int
	err := rows.Scan(&id, &item.Png, &item.Link, &item.Redirect)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return item, false
		}
		log.Printf("Error scanning for item: %s", err.Error())
		return item, false
	}
	return item, true

}

// get an asset from the store
func (s *SQLiteRepo) GetAsset(name string) (Asset, bool) {
	rows := s.db.QueryRow("SELECT * FROM assets WHERE name = ?", name)
	var item Asset
	var id int
	err := rows.Scan(&id, &item.Name, &item.Data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return item, false
		}
		log.Printf("Error getting asset: %s", err.Error())
		return item, false
	}
	return item, true

}

/*
Get all nav bar items. Returns a list of NavBarItem structs with the png data, the file name, and the redirect location of the icon

/*
Get all nav bar items. Returns a list of NavBarItem structs with the png data, the file name, and the redirect location of the icon
*/
func (s *SQLiteRepo) GetNavBarLinks() []NavBarItem {
	items, err := s.GetNavBarLinksWithError()
	if err != nil {
		log.Printf("get navbar links: %v", err)
	}
	return items
}

func (s *SQLiteRepo) GetNavBarLinksWithError() ([]NavBarItem, error) {
	rows, err := s.db.Query("SELECT * FROM navbar")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []NavBarItem
	for rows.Next() {
		var item NavBarItem
		var id int
		if err := rows.Scan(&id, &item.Png, &item.Link, &item.Redirect); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *SQLiteRepo) GetAssets() []Asset {
	items, err := s.GetAssetsWithError()
	if err != nil {
		log.Printf("get assets: %v", err)
	}
	return items
}

func (s *SQLiteRepo) GetAssetsWithError() ([]Asset, error) {
	rows, err := s.db.Query("SELECT * FROM assets")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Asset
	for rows.Next() {
		var item Asset
		var id int
		if err := rows.Scan(&id, &item.Name, &item.Data); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *SQLiteRepo) GetAdminTables() AdminPage {
	page, err := s.GetAdminTablesWithError()
	if err != nil {
		log.Printf("get admin tables: %v", err)
	}
	return page
}

func (s *SQLiteRepo) GetAdminTablesWithError() (AdminPage, error) {
	rows, err := s.db.Query("SELECT * FROM admin")
	if err != nil {
		return AdminPage{}, err
	}
	defer rows.Close()
	page := AdminPage{Tables: map[string][]TableData{}}
	for rows.Next() {
		var item TableData
		var id int
		var category string
		if err := rows.Scan(&id, &item.DisplayName, &item.Link, &category); err != nil {
			return AdminPage{}, err
		}
		page.Tables[category] = append(page.Tables[category], item)
	}
	if err := rows.Err(); err != nil {
		return AdminPage{}, err
	}
	return page, nil
}

/*
Retrieve a document from the sqlite db

	:param id: the Identifier of the post
*/
func (s *SQLiteRepo) GetDocument(id Identifier) (Document, error) {
	row := s.db.QueryRow("SELECT * FROM posts WHERE id = ?", id)

	var post Document
	var rowNum int
	if err := row.Scan(&rowNum, &post.Ident, &post.Title, &post.Created, &post.Body, &post.Category, &post.Sample); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return post, ErrNotExists
		}
		return post, err
	}
	return post, nil

}

/*
Get all documents by category

	:param category: the category to retrieve all docs from
*/
func (s *SQLiteRepo) GetByCategory(category string) []Document {
	docs, err := s.GetByCategoryWithError(category)
	if err != nil {
		log.Printf("get documents by category: %v", err)
	}
	return docs
}

func (s *SQLiteRepo) GetByCategoryWithError(category string) ([]Document, error) {
	return s.listDocuments("SELECT * FROM posts WHERE category = ?", category)
}

func (s *SQLiteRepo) listDocuments(query string, args ...interface{}) ([]Document, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var docs []Document
	for rows.Next() {
		var doc Document
		if err := rows.Scan(&doc.Row, &doc.Ident, &doc.Title, &doc.Created, &doc.Body, &doc.Category, &doc.Sample); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return docs, nil
}

/*
get image data from the images table

	:param id: the serial identifier of the post
*/
func (s *SQLiteRepo) imageMetadata(id Identifier) (Image, error) {
	if err := validateImageIdentifier(id); err != nil {
		return Image{}, ErrNotExists
	}
	row := s.db.QueryRow("SELECT id, title, desc, created, category, COALESCE(media_type, '') FROM images WHERE id = ?", id)
	var img Image
	if err := row.Scan(&img.Ident, &img.Title, &img.Desc, &img.Created, &img.Category, &img.MediaType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Image{}, ErrNotExists
		}
		return Image{}, err
	}
	return img, nil
}

func (s *SQLiteRepo) GetImage(id Identifier) (Image, error) {
	img, err := s.imageMetadata(id)
	if err != nil {
		return Image{}, err
	}
	img.Data, err = s.imageIO.Get(id)
	if err != nil {
		return Image{}, err
	}
	return img, nil
}

type memoryImage struct{ *bytes.Reader }

func (memoryImage) Close() error { return nil }

// OpenImage returns metadata and a seekable stream. Legacy ImageIO implementations
// remain usable through a buffered fallback; callers must close the stream.
func (s *SQLiteRepo) OpenImage(id Identifier) (Image, ReadSeekCloser, error) {
	img, err := s.imageMetadata(id)
	if err != nil {
		return Image{}, nil, err
	}
	if opener, ok := s.imageIO.(ImageOpener); ok {
		file, err := opener.Open(id)
		if err != nil {
			return Image{}, nil, err
		}
		return img, file, nil
	}
	data, err := s.imageIO.Get(id)
	if err != nil {
		return Image{}, nil, err
	}
	return img, memoryImage{bytes.NewReader(data)}, nil
}

/*
Get all of the images from the datastore
*/
func (s *SQLiteRepo) GetAllImages() []Image {
	imgs, err := s.GetAllImagesWithError()
	if err != nil {
		log.Printf("get all images: %v", err)
	}
	return imgs
}

func (s *SQLiteRepo) GetAllImagesWithError() ([]Image, error) {
	return s.listImages("SELECT id, title, desc, created, category, COALESCE(media_type, '') FROM images")
}

func (s *SQLiteRepo) GetImagesByCategory(category string) []Image {
	imgs, err := s.GetImagesByCategoryWithError(category)
	if err != nil {
		log.Printf("get images by category: %v", err)
	}
	return imgs
}

func (s *SQLiteRepo) GetImagesByCategoryWithError(category string) ([]Image, error) {
	return s.listImages("SELECT id, title, desc, created, category, COALESCE(media_type, '') FROM images WHERE category = ?", category)
}

// listImages deliberately returns metadata only, even if a media file is missing.
func (s *SQLiteRepo) listImages(query string, args ...interface{}) ([]Image, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	imgs := []Image{}
	for rows.Next() {
		var img Image
		if err := rows.Scan(&img.Ident, &img.Title, &img.Desc, &img.Created, &img.Category, &img.MediaType); err != nil {
			return nil, err
		}
		imgs = append(imgs, img)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return imgs, nil
}

/*
Add an image to the database

	:param title: the title of the image
	:param location: the location to save the image to
	:param desc: the description of the image, if any
	:param data: the binary data for the image
*/
func (s *SQLiteRepo) AddImage(data []byte, title, desc, category, mediaType string) (Identifier, error) {
	id := newIdentifier()
	err := s.imageIO.Put(data, id)
	if err != nil {
		return Identifier(""), err
	}
	_, err = s.db.Exec("INSERT INTO images (id, title, desc, created, category, media_type) VALUES (?,?,?,?,?,?)", string(id), title, desc, time.Now().String(), category, mediaType)
	if err != nil {
		return Identifier(""), err
	}
	return id, nil
}

// UpdateDigitalArt changes only the title and description of a digital art
// record. Other image categories and the image file are never modified.
func (s *SQLiteRepo) UpdateDigitalArt(id Identifier, title, desc string) error {
	if err := validateImageIdentifier(id); err != nil {
		return err
	}
	res, err := s.db.Exec("UPDATE images SET title = ?, desc = ? WHERE id = ? AND category = ? AND (SELECT COUNT(*) FROM images WHERE id = ?) = 1", title, desc, id, DIGITAL_ART, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrNotExists
	}
	return nil
}

// ArchiveDigitalArt removes an art listing while retaining its media and ID for
// existing references. Ambiguous IDs are left untouched, regardless of category.
func (s *SQLiteRepo) ArchiveDigitalArt(id Identifier) error {
	if err := validateImageIdentifier(id); err != nil {
		return err
	}
	res, err := s.db.Exec("UPDATE images SET category = ? WHERE id = ? AND category = ? AND (SELECT COUNT(*) FROM images WHERE id = ?) = 1", MISC_ASSET, id, DIGITAL_ART, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrNotExists
	}
	return nil
}

/*
Updates a document in the database with the supplied. Only changes the title, the body, category. Keys off of the documents Identifier

	:param doc: the Document to upload into the database
*/
func (s *SQLiteRepo) UpdateDocument(doc Document) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("UPDATE posts SET title = ?, body = ?, category = ?, sample = ? WHERE id = ?;")
	if err != nil {
		tx.Rollback()
		return err
	}

	res, err := stmt.Exec(doc.Title, doc.Body, doc.Category, doc.MakeSample(), doc.Ident)
	if err != nil {
		tx.Rollback()
		return err
	}
	affected, _ := res.RowsAffected()
	if affected != 1 {
		return ErrNotExists
	}

	tx.Commit()
	return nil
}

/*
Adds a LinkPair to the menu database table

	:param item: the LinkPair to upload
*/
func (s *SQLiteRepo) AddMenuItem(item LinkPair) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	fmt.Printf("from AddMenuItem: %+v\n", item)
	foundItem, found := s.GetMenuItemByName(item.Link, item.Text)
	if found {
		fmt.Printf("from error in AddMenuItem: %+v\n", foundItem)
		tx.Rollback()
		return ErrDuplicate
	}
	stmt, _ := tx.Prepare("INSERT INTO menu(link, text) VALUES (?,?)")
	_, err = stmt.Exec(item.Link, item.Text)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil

}

/*
Adds an item to the navbar database table

	:param item: the NavBarItem to upload
*/
func (s *SQLiteRepo) AddNavbarItem(item NavBarItem) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	_, found := s.GetNavbarLink(item.Link, item.Redirect)
	if found {
		tx.Rollback()
		return ErrDuplicate
	}
	stmt, err := tx.Prepare("INSERT INTO navbar(png, link, redirect) VALUES (?,?,?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = stmt.Exec(item.Png, item.Link, item.Redirect)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil

}

/*
Adds an asset to the asset database table asset

	:param name: the name of the asset (filename)
	:param data: the byte array of the PNG to upload TODO: limit this to 256kb
*/
func (s *SQLiteRepo) AddAsset(name string, data []byte) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	_, found := s.GetAsset(name)
	if found {
		tx.Rollback()
		return ErrDuplicate
	}
	stmt, _ := tx.Prepare("INSERT INTO assets(name, data) VALUES (?,?)")
	_, err = stmt.Exec(name, data)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

/*
Adds a document to the database (for text posts)

	:param doc: the Document to add
*/
func (s *SQLiteRepo) AddDocument(doc Document) (Identifier, error) {
	id := newIdentifier()
	tx, err := s.db.Begin()
	if err != nil {
		return Identifier(""), err
	}
	stmt, _ := tx.Prepare("INSERT INTO posts(id, title, created, body, category, sample) VALUES (?,?,?,?,?,?)")
	_, err = stmt.Exec(id, doc.Title, doc.Created, doc.Body, doc.Category, doc.MakeSample())
	if err != nil {
		tx.Rollback()
		return Identifier(""), err
	}
	tx.Commit()
	return id, nil

}

/*
Add an entry to the 'admin' table in the database

	:param item: an admin table k/v text to redirect pair
	:param tableName: the name of the table to populate the link in on the UI
*/
func (s *SQLiteRepo) AddAdminTableEntry(item TableData, category string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	_, found := s.GetAdminTableEntry(item.DisplayName, item.Link, category)
	if found {
		tx.Rollback()
		return ErrDuplicate
	}
	stmt, _ := tx.Prepare("INSERT INTO admin (display_name, link, category) VALUES (?,?,?)")
	_, err = stmt.Exec(item.DisplayName, item.Link, category)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	return nil
}

/*
Delete a document from the db

	:param id: the identifier of the document to remove
*/
func (s *SQLiteRepo) DeleteDocument(id Identifier) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, _ := tx.Prepare("DELETE FROM posts WHERE id=?")
	_, err = stmt.Exec(id)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil

}

/*
Delete navigation bar item

	:param id: the name or 'id' of the navbar item to remove
*/
func (s *SQLiteRepo) DeleteNavbarItem(id Identifier) error {

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, _ := tx.Prepare("DELETE FROM navbar WHERE link=?")
	result, err := stmt.Exec(id)
	if err != nil {
		tx.Rollback()
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected < 1 {
		return ErrNotExists

	}

	tx.Commit()
	return nil

}

// Get all Hosts from the host table
func (s *SQLiteRepo) AllDocuments() []Document {
	docs, err := s.AllDocumentsWithError()
	if err != nil {
		log.Printf("get all documents: %v", err)
	}
	return docs
}

func (s *SQLiteRepo) AllDocumentsWithError() ([]Document, error) {
	return s.listDocuments("SELECT * FROM posts")
}

type InvalidSkipArg struct{ Skip int }

func (i *InvalidSkipArg) Error() string {
	return fmt.Sprintf("Invalid skip amount was passed: %v", i.Skip)
}

type ImageStoreItem struct {
	Identifier   string `json:"identifier"`
	Filename     string `json:"filename"`
	AbsolutePath string `json:"absolute_path"`
	Title        string `json:"title" form:"title"`
	Created      string `json:"created"`
	Desc         string `json:"description" form:"description"`
	Category     string `json:"category"`
	ApiPath      string
}

/*
Function to return the location of the image store. Wrapping the env call in
a function so that refactoring is easier
*/
func GetImageStore() string {
	return os.Getenv(env.IMAGE_STORE)
}

// Wrapping the new id call in a function to make refactoring easier
func newIdentifier() Identifier {
	return Identifier(uuid.NewString())
}
