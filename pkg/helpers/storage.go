package helpers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"git.aetherial.dev/aeth/keiji/pkg/env"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
)

type DatabaseSchema struct {
	// Gotta figure out what this looks like
	// so that the ExtractAll() function gets
	// all of the data from the database

}

type MenuLinkPair struct {
	MenuLink string `json:"menu_link"`
	LinkText string `json:"link_text"`
}

type NavBarItem struct {
	Png  string `json:"png"`
	Link string `json:"link"`
}

type Document struct {
	PostId   string
	Title    string `json:"title"`
	Created  string `json:"created"`
	Body     string `json:"body"`
	Category string `json:"category"`
	Sample   string
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
	Location string
	Title    string
	Desc     string
}

type DocumentIO interface {
	GetDocument(id string) (Document, error)
	GetImage(id string) (Image, error)
	UpdateDocument(doc Document) error
	DeleteDocument(id string) error
	AddDocument(doc Document) error
	AddImage(img ImageStoreItem) error
	GetByCategory(category string) []string
	AllDocuments() []Document
	GetDropdownElements() []MenuLinkPair
	GetNavBarLinks() []NavBarItem
	ExportAll()
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
	query := `
    CREATE TABLE IF NOT EXISTS posts(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		postid TEXT NOT NULL,
		title TEXT NOT NULL,
        created TEXT NOT NULL,
        body TEXT NOT NULL UNIQUE,
        category TEXT NOT NULL,
		sample TEXT NOT NULL
    );
    `

	_, err := r.db.Exec(query)
	return err
}

/*
Create an entry in the hosts table

	:param host: a Host entry from a port scan
*/
func (r *SQLiteRepo) Create(post Document) error {
	_, err := r.db.Exec("INSERT INTO posts(postid, title, created, body, category, sample) values(?,?,?,?,?)", uuid.New().String(), post.Title, post.Created, post.Body, post.Category, post.MakeSample())
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return ErrDuplicate
			}
		}
		return err
	}

	return nil
}

// Get all Hosts from the host table
func (r *SQLiteRepo) AllDocuments() []Document {
	rows, err := r.db.Query("SELECT * FROM posts")
	if err != nil {
		fmt.Printf("There was an issue getting all posts. %s", err.Error())
		return nil
	}
	defer rows.Close()

	var all []Document
	for rows.Next() {
		var post Document
		if err := rows.Scan(&post.PostId, &post.Title, &post.Created, &post.Body, &post.Sample); err != nil {
			fmt.Printf("There was an error getting all documents. %s", err.Error())
			return nil
		}
		all = append(all, post)
	}
	return all
}

// Get a blogpost by its postid
func (r *SQLiteRepo) GetByIP(postId string) (Document, error) {
	row := r.db.QueryRow("SELECT * FROM posts WHERE postid = ?", postId)

	var post Document
	if err := row.Scan(&post); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return post, ErrNotExists
		}
		return post, err
	}
	return post, nil
}

// Update a record by its ID
func (r *SQLiteRepo) Update(id int64, updated Document) error {
	if id == 0 {
		return errors.New("invalid updated ID")
	}
	res, err := r.db.Exec("UPDATE posts SET title = ?, body = ?, desc = ? WHERE id = ?", updated.Title, updated.Body, updated.MakeSample(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUpdateFailed
	}

	return nil
}

// Delete a record by its ID
func (r *SQLiteRepo) Delete(id int64) error {
	res, err := r.db.Exec("DELETE FROM posts WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrDeleteFailed
	}

	return err
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
func NewImageStoreItem(fname string, title string, desc string) *ImageStoreItem {
	id := uuid.New()
	img := ImageStoreItem{
		Identifier:   id.String(),
		Filename:     fname,
		Title:        title,
		Category:     DIGITAL_ART,
		AbsolutePath: fmt.Sprintf("%s/%s", GetImageStore(), fname),
		Created:      time.Now().UTC().String(),
		Desc:         desc,
	}
	return &img
}

/*
Function to return the location of the image store. Wrapping the env call in
a function so that refactoring is easier
*/
func GetImageStore() string {
	return os.Getenv(env.IMAGE_STORE)
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
