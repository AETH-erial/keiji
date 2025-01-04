package storage

import (
	"database/sql"
	"log"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func newTestDb() (*SQLiteRepo, *sql.DB) {

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	testDb := &SQLiteRepo{db: db}
	err = testDb.Migrate()
	if err != nil {
		log.Fatal("failed to start the test database: ", err)
	}
	return testDb, db
}

func TestMigrate(t *testing.T) {
	requiredTables := []string{
		"posts",
		"images",
		"menu",
		"navbar",
		"assets",
		"admin",
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	testDb := &SQLiteRepo{db: db}
	err = testDb.Migrate()
	if err != nil {
		t.Error(err)
	}
	for i := range requiredTables {
		name := requiredTables[i]
		row := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='?'", name)
		if row.Err() != nil {
			t.Errorf("error querying table: %s", name)
		}
		if row == nil {
			t.Errorf("no table returned: %s", name)
		}
	}

}

func TestGetDropdownElements(t *testing.T) {
	type testcase struct {
		seed []LinkPair
	}
	testDb, db := newTestDb()
	for _, tc := range []testcase{
		{
			seed: []LinkPair{
				{
					Text: "abc123",
					Link: "/abc/123",
				},
			},
		},
	} {
		stmt, _ := db.Prepare("INSERT INTO menu(link, text) VALUES (?,?)")
		for i := range tc.seed {
			_, err := stmt.Exec(tc.seed[i].Link, tc.seed[i].Text)
			if err != nil {
				t.Errorf("failed to seed: %s", err)
			}
		}
		got := testDb.GetDropdownElements()
		assert.Equal(t, got, tc.seed)

	}

}
func TestGetNavBarLinks(t *testing.T) {
	type testcase struct {
		seed []NavBarItem
	}
	testDb, db := newTestDb()
	for _, tc := range []testcase{
		{
			seed: []NavBarItem{
				{
					Link:     "/abc/123",
					Redirect: "/abc/123/site",
					Png:      []byte("xzy123abc098"),
				},
			},
		},
	} {
		stmt, _ := db.Prepare("INSERT INTO navbar(png, link, redirect) VALUES (?,?,?)")
		for i := range tc.seed {
			_, err := stmt.Exec(tc.seed[i].Png, tc.seed[i].Link, tc.seed[i].Redirect)
			if err != nil {
				t.Errorf("failed to seed: %s", err)
			}
		}
		got := testDb.GetNavBarLinks()
		assert.Equal(t, tc.seed, got)

	}
}
func TestGetAssets(t *testing.T) {
	type testcase struct {
		seed []Asset
	}
	testDb, db := newTestDb()
	for _, tc := range []testcase{
		{
			seed: []Asset{
				{
					Data: []byte("abc123xyz098"),
					Name: "asset1",
				},
			},
		},
	} {
		stmt, _ := db.Prepare("INSERT INTO assets(data, name) VALUES (?,?)")
		for i := range tc.seed {
			_, err := stmt.Exec(tc.seed[i].Data, tc.seed[i].Name)
			if err != nil {
				t.Error(err)
			}
		}
		got := testDb.GetAssets()
		assert.Equal(t, tc.seed, got)

	}
}

func TestGetAdminTables(t *testing.T) {

	type testcase struct {
		seed AdminPage
	}
	testDb, db := newTestDb()
	for _, tc := range []testcase{
		{
			seed: AdminPage{
				Tables: map[string][]TableData{
					"test": {
						{
							DisplayName: "abc123",
							Link:        "xyz098",
						},
					},
				},
			},
		},
	} {
		stmt, _ := db.Prepare("INSERT INTO admin(display_name, link, category) VALUES (?,?,?)")
		for k, table := range tc.seed.Tables {
			for i := range table {
				_, err := stmt.Exec(table[i].DisplayName, table[i].Link, k)
				if err != nil {
					t.Error(err)
				}

			}
		}
		got := testDb.GetAdminTables()
		assert.Equal(t, tc.seed, got)

	}
}
func TestGetDocument(t *testing.T) {

	type testcase struct {
		seed Document
	}
	testDb, db := newTestDb()
	for _, tc := range []testcase{
		{
			seed: Document{
				Ident:    Identifier("qwerty"),
				Title:    "abc 123",
				Created:  "2024-12-31",
				Body:     "blog post body etc",
				Category: BLOG,
				Sample:   "this is a sample",
			},
		},
	} {
		stmt, _ := db.Prepare("INSERT INTO posts(id, title, created, body, category, sample) VALUES (?,?,?,?,?,?)")
		_, err := stmt.Exec(tc.seed.Ident, tc.seed.Title, tc.seed.Created, tc.seed.Body, tc.seed.Category, tc.seed.Sample)
		if err != nil {
			t.Error(err)
		}
		got, _ := testDb.GetDocument(Identifier("qwerty"))
		assert.Equal(t, tc.seed, got)

	}
}
func TestGetByCategory(t *testing.T) {

	type testcase struct {
		seed []Document
	}
	testDb, db := newTestDb()
	for _, tc := range []testcase{
		{
			seed: []Document{
				{
					Row:      1,
					Ident:    Identifier("qwerty"),
					Title:    "abc 123",
					Created:  "2024-12-31",
					Body:     "blog post body etc",
					Category: BLOG,
					Sample:   "this is a sample",
				},
				{
					Row:      2,
					Ident:    Identifier("poiuyt"),
					Title:    "abc 123",
					Created:  "2024-12-31",
					Body:     "blog post body etc",
					Category: BLOG,
					Sample:   "this is a sample",
				},
			},
		},
	} {
		stmt, _ := db.Prepare("INSERT INTO posts(id, title, created, body, category, sample) VALUES (?,?,?,?,?,?)")
		for i := range tc.seed {
			_, err := stmt.Exec(tc.seed[i].Ident, tc.seed[i].Title, tc.seed[i].Created, tc.seed[i].Body, tc.seed[i].Category, tc.seed[i].Sample)
			if err != nil {
				t.Error(err)
			}
		}
		got := testDb.GetByCategory(BLOG)
		assert.Equal(t, tc.seed, got)
	}

}
func TestGetImage(t *testing.T) {

	// testDb, db := newTestDb()
}
func TestGetAllImages(t *testing.T) {

	// testDb, db := newTestDb()
}

func TestAllDocuments(t *testing.T) {
	// testDb, db := newTestDb()
}

func TestUpdateDocument(t *testing.T) {

	// testDb, db := newTestDb()
}

func TestAddImage(t *testing.T) {

	// testDb, db := newTestDb()
}
func TestAddMenuItem(t *testing.T) {

	// testDb, db := newTestDb()
}
func TestAddNavbarItem(t *testing.T) {

	// testDb, db := newTestDb()
}
func TestAddAsset(t *testing.T) {

	// testDb, db := newTestDb()
}
func TestAddDocument(t *testing.T) {

	// testDb, db := newTestDb()
}
func TestAddAdminTableEntry(t *testing.T) {

	// testDb, db := newTestDb()
}

func TestDeleteDocument(t *testing.T) {

	// testDb, db := newTestDb()
}

func TestGetImageStore(t *testing.T) {

	// testDb, db := newTestDb()
}
func TestNewIdentifier(t *testing.T) {

	// testDb, db := newTestDb()
}
