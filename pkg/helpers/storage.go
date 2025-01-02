package helpers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"time"

	"git.aetherial.dev/aeth/keiji/pkg/env"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type DatabaseSchema struct {
	// Gotta figure out what this looks like
	// so that the ExtractAll() function gets
	// all of the data from the database

}

type MenuLinkPair struct {
	MenuLink string `json:"link"`
	LinkText string `json:"text"`
}

type NavBarItem struct {
	Png      []byte `json:"png"`
	Link     string `json:"link"`
	Redirect string `json:"redirect"`
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
	Ident    Identifier            `json:"identifier"`
	Location string                `json:"title" form:"title"`
	Title    string                `json:"description" form:"description"`
	File     *multipart.FileHeader `form:"file"`
	Desc     string
	Created  string
	Category string
	Data     []byte
}

type DocumentIO interface {
	GetDocument(id Identifier) (Document, error)
	GetImage(id Identifier) (Image, error)
	GetAllImages() []Image
	UpdateDocument(doc Document) error
	DeleteDocument(id Identifier) error
	AddDocument(doc Document) error
	AddImage(data []byte, title, desc string) error
	AddAdminTableEntry(TableData, string) error
	AddNavbarItem(NavBarItem) error
	AddMenuItem(MenuLinkPair) error
	GetByCategory(category string) []Document
	AllDocuments() []Document
	GetDropdownElements() []MenuLinkPair
	GetNavBarLinks() []NavBarItem
	GetAssets() []Asset
	GetAdminTables() AdminPage
}

var (
	ErrDuplicate    = errors.New("record already exists")
	ErrNotExists    = errors.New("row not exists")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)

type SQLiteRepo struct {
	db *sql.DB
}

// Instantiate a new SQLiteRepo struct
func NewSQLiteRepo(db *sql.DB) *SQLiteRepo {
	return &SQLiteRepo{
		db: db,
	}

}

// Creates a new SQL table for text posts
func (r *SQLiteRepo) Migrate() error {
	postsTable := `
    CREATE TABLE IF NOT EXISTS posts(
        row INTEGER PRIMARY KEY AUTOINCREMENT,
		id TEXT NOT NULL UNIQUE,
		title TEXT NOT NULL,
        created TEXT NOT NULL,
        body TEXT NOT NULL,
        category TEXT NOT NULL,
		sample TEXT NOT NULL
    );
    `
	imagesTable := `
	CREATE TABLE IF NOT EXISTS images(
		row INTEGER PRIMARY KEY AUTOINCREMENT,
		id TEXT NOT NULL,
		title TEXT NOT NULL,
		location TEXT NOT NULL,
		desc TEXT NOT NULL,
		created TEXT NOT NULL
	);
	`
	menuItemsTable := `
	CREATE TABLE IF NOT EXISTS menu(
		row INTEGER PRIMARY KEY AUTOINCREMENT,
		link TEXT NOT NULL,
		text TEXT NOT NULL
	);
	`
	navbarItemsTable := `
	CREATE TABLE IF NOT EXISTS navbar(
		row INTEGER PRIMARY KEY AUTOINCREMENT,
		png BLOB NOT NULL,
		link TEXT NOT NULL,
		redirect TEXT
	);`
	assetTable := `
	CREATE TABLE IF NOT EXISTS assets(
		row INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		data BLOB NOT NULL
	);
	`
	adminTable := `
	CREATE TABLE IF NOT EXISTS admin(
		row INTEGER PRIMARY KEY AUTOINCREMENT,
		display_name TEXT NOT NULL,
		link TEXT NOT NULL,
		category TEXT NOT NULL
	);
	`
	seedQueries := []string{postsTable, imagesTable, menuItemsTable, navbarItemsTable, assetTable, adminTable}
	for i := range seedQueries {
		_, err := r.db.Exec(seedQueries[i])
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Seed the database with the necessary configuration items to function properly

	:param menu: the text file containing the k/v pair for the navigation menu
	:param pngs: the text file containing the k/v pair for the icon names -> redirect links
	:param dir: the directory that the PNG assets are in (note: the k/v pair in pngs will read from this dir)
*/
func (s *SQLiteRepo) Seed(menu string, pngs string, dir string) { // TODO: make a bootstrap file with a comprehensive unmarshalling sequence for tighter control of the seeing procedute
	b, err := os.ReadFile(menu)

	if err != nil {
		log.Fatal(err)
	}
	entries := strings.Split(string(b), "\n")
	for i := range entries {
		if entries[i] == "" {
			continue
		}
		info := strings.Split(entries[i], "=")
		err := s.AddMenuItem(MenuLinkPair{MenuLink: info[0], LinkText: info[1]})
		if err != nil {
			log.Fatal(err)
		}
	}
	b, err = os.ReadFile(pngs)
	if err != nil {
		log.Fatal(err)
	}
	entries = strings.Split(string(b), "\n")
	for i := range entries {
		if entries[i] == "" {
			continue
		}
		info := strings.Split(entries[i], "=")
		b, err := os.ReadFile(path.Join(dir, info[0]))
		if err != nil {
			log.Fatal(err)
		}
		err = s.AddNavbarItem(NavBarItem{Png: b, Link: info[0], Redirect: info[1]})
		if err != nil {
			log.Fatal(err)
		}
	}
	assets, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for i := range assets {
		b, err := os.ReadFile(path.Join(dir, assets[i].Name()))
		if err != nil {
			log.Fatal(err)
		}
		err = s.AddAsset(assets[i].Name(), b)
		if err != nil {
			log.Fatal(err)
		}

	}

}

/*
Get all dropdown menu elements. Returns a list of MenuLinkPair structs with the text and redirect location
*/
func (s *SQLiteRepo) GetDropdownElements() []MenuLinkPair {
	rows, err := s.db.Query("SELECT * FROM menu")
	var menuItems []MenuLinkPair
	defer rows.Close()
	for rows.Next() {
		var id int
		var item MenuLinkPair
		err = rows.Scan(&id, &item.MenuLink, &item.LinkText)
		if err != nil {
			log.Fatal(err)
		}
		menuItems = append(menuItems, item)
	}
	return menuItems

}

/*
Get all nav bar items. Returns a list of NavBarItem structs with the png data, the file name, and the redirect location of the icon
*/
func (s *SQLiteRepo) GetNavBarLinks() []NavBarItem {

	rows, err := s.db.Query("SELECT * FROM navbar")
	var navbarItems []NavBarItem
	defer rows.Close()
	for rows.Next() {
		var item NavBarItem
		var id int
		err = rows.Scan(&id, &item.Png, &item.Link, &item.Redirect)
		if err != nil {
			log.Fatal(err)
		}
		navbarItems = append(navbarItems, item)
	}
	return navbarItems

}

/*
get all assets from the asset table
*/
func (s *SQLiteRepo) GetAssets() []Asset {
	rows, err := s.db.Query("SELECT * FROM assets")
	var assets []Asset
	defer rows.Close()
	for rows.Next() {
		var item Asset
		var id int
		err = rows.Scan(&id, &item.Name, &item.Data)
		if err != nil {
			log.Fatal(err)
		}
		assets = append(assets, item)
	}
	return assets

}

/*
get all assets from the asset table
*/
func (s *SQLiteRepo) GetAdminTables() AdminPage {
	rows, err := s.db.Query("SELECT * FROM admin")
	adminPage := AdminPage{Tables: map[string][]TableData{}}
	defer rows.Close()
	for rows.Next() {
		var item TableData
		var id int
		var category string
		err = rows.Scan(&id, &item.DisplayName, &item.Link, &category)
		if err != nil {
			log.Fatal(err)
		}
		adminPage.Tables[category] = append(adminPage.Tables[category], item)
	}
	return adminPage

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
	rows, err := s.db.Query("SELECT * FROM posts WHERE category = ?", category)
	if err != nil {
		log.Fatal(err)
	}
	var docs []Document
	defer rows.Close()
	for rows.Next() {
		var doc Document
		err := rows.Scan(&doc.Row, &doc.Ident, &doc.Title, &doc.Created, &doc.Body, &doc.Category, &doc.Sample)
		if err != nil {
			log.Fatal(err)
		}
		docs = append(docs, doc)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return docs

}

/*
get image data from the images table

	:param id: the serial identifier of the post
*/
func (s *SQLiteRepo) GetImage(id Identifier) (Image, error) {
	row := s.db.QueryRow("SELECT * FROM images WHERE id = ?", id)
	var rowNum int
	var title string
	var location string
	var desc string
	var created string
	if err := row.Scan(&rowNum, &title, &location, &desc, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Image{}, ErrNotExists
		}
		return Image{}, err
	}
	data, err := os.ReadFile(location)
	if err != nil {
		return Image{}, err
	}
	return Image{Ident: id, Title: title, Location: location, Desc: desc, Data: data, Created: created}, nil
}

/*
Get all of the images from the datastore
*/
func (s *SQLiteRepo) GetAllImages() []Image {
	rows, err := s.db.Query("SELECT * FROM images")
	if err != nil {
		log.Fatal(err)
	}
	imgs := []Image{}
	for rows.Next() {
		var img Image
		var rowNum int
		err := rows.Scan(&rowNum, &img.Ident, &img.Title, &img.Location, &img.Desc, &img.Created)
		if err != nil {
			log.Fatal(err)
		}
		b, err := os.ReadFile(img.Location)
		if err != nil {
			log.Fatal(err)
		}
		imgs = append(imgs, Image{Ident: img.Ident, Title: img.Title, Location: img.Location, Desc: img.Desc, Data: b, Created: img.Created})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return imgs
}

/*
Add an image to the database

	:param title: the title of the image
	:param location: the location to save the image to
	:param desc: the description of the image, if any
	:param data: the binary data for the image
*/
func (s *SQLiteRepo) AddImage(data []byte, title string, desc string) error {
	id := newIdentifier()
	fsLoc := path.Join(GetImageStore(), string(id))
	err := os.WriteFile(fsLoc, data, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO images (id, title, location, desc, created) VALUES (?,?,?,?,?)", string(id), title, fsLoc, desc, time.Now().String())
	if err != nil {
		return err
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
	_, err = stmt.Exec(doc.Title, doc.Body, doc.Category, doc.MakeSample(), doc.Ident)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

/*
Adds a MenuLinkPair to the menu database table

	:param item: the MenuLinkPair to upload
*/
func (s *SQLiteRepo) AddMenuItem(item MenuLinkPair) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, _ := tx.Prepare("INSERT INTO menu(link, text) VALUES (?,?)")
	_, err = stmt.Exec(item.MenuLink, item.LinkText)
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
func (s *SQLiteRepo) AddDocument(doc Document) error {
	id := uuid.New()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, _ := tx.Prepare("INSERT INTO posts(id, title, created, body, category, sample) VALUES (?,?,?,?,?,?)")
	_, err = stmt.Exec(id.String(), doc.Title, doc.Created, doc.Body, doc.Category, doc.MakeSample())
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil

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

// Get all Hosts from the host table
func (s *SQLiteRepo) AllDocuments() []Document {
	rows, err := s.db.Query("SELECT * FROM posts")
	if err != nil {
		fmt.Printf("There was an issue getting all posts. %s", err.Error())
		return nil
	}
	defer rows.Close()

	all := []Document{}
	for rows.Next() {
		var post Document
		if err := rows.Scan(&post.Ident, &post.Title, &post.Created, &post.Body, &post.Sample); err != nil {
			fmt.Printf("There was an error getting all documents. %s", err.Error())
			return nil
		}
		all = append(all, post)
	}
	return all
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
Create a new ImageStoreItem

	:param fname: the name of the file to be saved
	:param title: the canonical title to give the image
	:param desc: the description to associate to the image
*/
func NewImageStoreItem(title string, desc string) Image {
	id := newIdentifier()
	img := Image{
		Ident:    id,
		Title:    title,
		Category: DIGITAL_ART,
		Location: GetImageStore(),
		Created:  time.Now().UTC().String(),
		Desc:     desc,
	}
	return img
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

/*
Return database entries of the images that exist in the imagestore

	:param rds: pointer to a RedisCaller to perform the lookups with
*/
func GetImageData(rds *RedisCaller) ([]*ImageStoreItem, error) {
	ids, err := rds.GetByCategory(DIGITAL_ART)
	if err != nil {
		return nil, err
	}

	var imageEntries []*ImageStoreItem
	for i := range ids {
		val, err := rds.Client.Get(rds.ctx, ids[i]).Result()
		if err == redis.Nil {
			return nil, err
		} else if err != nil {
			return nil, err
		}
		data := []byte(val)
		var imageEntry ImageStoreItem
		err = json.Unmarshal(data, &imageEntry)
		if err != nil {
			return nil, err
		}
		imageEntry.ApiPath = fmt.Sprintf("/api/v1/images/%s", imageEntry.Filename)
		imageEntries = append(imageEntries, &imageEntry)
	}
	return imageEntries, err
}
