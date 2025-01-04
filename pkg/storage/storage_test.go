package storage

import (
	"database/sql"
	"errors"
	"log"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

const badPostsTable = `
    CREATE TABLE IF NOT EXISTS posts(
        row INTEGER PRIMARY KEY AUTOINCREMENT
    );
    `
const badImagesTable = `
	CREATE TABLE IF NOT EXISTS images(
		row INTEGER PRIMARY KEY AUTOINCREMENT
	);
	`
const badMenuItemsTable = `
	CREATE TABLE IF NOT EXISTS menu(
		row INTEGER PRIMARY KEY AUTOINCREMENT
	);
	`
const badNavbarItemsTable = `
	CREATE TABLE IF NOT EXISTS navbar(
		row INTEGER PRIMARY KEY AUTOINCREMENT
	);`
const badAssetTable = `
	CREATE TABLE IF NOT EXISTS assets(
		row INTEGER PRIMARY KEY AUTOINCREMENT
	);
	`
const badAdminTable = `
	CREATE TABLE IF NOT EXISTS admin(
		row INTEGER PRIMARY KEY AUTOINCREMENT
	);
	`

var unpopulatedTables = []string{badPostsTable, badImagesTable, badMenuItemsTable, badMenuItemsTable, badAssetTable, badAdminTable}

/*
creates in memory db and SQLiteRepo struct

	:param tmp: path to the temp directory for the filesystem IO struct to write images to
	:param migrate: choose to 'migrate' the database and create all the tables
*/
func newTestDb(tmp string, migrate bool) (*SQLiteRepo, *sql.DB) {

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	testDb := &SQLiteRepo{db: db, imageIO: FilesystemImageIO{RootDir: tmp}}
	if migrate {
		err = testDb.Migrate(RequiredTables)
	} else {
		err = testDb.Migrate(unpopulatedTables)
	}
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
	err = testDb.Migrate(RequiredTables)
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
	testDb, db := newTestDb(t.TempDir(), true)
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
	testDb, db := newTestDb(t.TempDir(), true)
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
	testDb, db := newTestDb(t.TempDir(), true)
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
	testDb, db := newTestDb(t.TempDir(), true)
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
	testDb, db := newTestDb(t.TempDir(), true)
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
	testDb, db := newTestDb(t.TempDir(), true)
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

	testDb, db := newTestDb(t.TempDir(), true)
	type testcase struct {
		seed       Image
		shouldSeed bool
		err        error
	}
	for _, tc := range []testcase{
		{
			seed: Image{
				Ident:   Identifier("abc123"),
				Title:   "xyz098",
				Desc:    "description",
				Created: "2024-12-31",
				Data:    []byte("abc123xyz098"),
			},
			shouldSeed: true,
			err:        nil,
		},
		{
			seed: Image{
				Ident: Identifier("zxcvbnm"),
			},
			shouldSeed: false,
			err:        ErrNotExists,
		},
	} {
		if tc.shouldSeed {
			_, err := db.Exec("INSERT INTO images (id, title, desc, created) VALUES (?,?,?,?)", string(tc.seed.Ident), tc.seed.Title, tc.seed.Desc, "2024-12-31")
			if err != nil {
				t.Error(err)
			}
			testDb.imageIO.Put(tc.seed.Data, tc.seed.Ident)
		}
		got, err := testDb.GetImage(tc.seed.Ident)
		if err != nil {
			assert.Equal(t, tc.err, err)
		} else {
			assert.Equal(t, tc.seed, got)
		}
	}
}
func TestGetAllImages(t *testing.T) {

	testDb, db := newTestDb(t.TempDir(), true)
	type testcase struct {
		seed []Image
	}
	for _, tc := range []testcase{
		{
			seed: []Image{
				{
					Ident:   Identifier("abc123"),
					Title:   "xyz098",
					Data:    []byte("abc123xyz098"),
					Created: "2024-12-31",
					Desc:    "description",
				},
				{
					Ident:   Identifier("xyz098"),
					Title:   "abc123",
					Data:    []byte("abc123xyz098"),
					Created: "2024-12-31",
					Desc:    "description",
				},
			},
		},
	} {
		for i := range tc.seed {
			_, err := db.Exec("INSERT INTO images (id, title, desc, created) VALUES (?,?,?,?)", string(tc.seed[i].Ident), tc.seed[i].Title, tc.seed[i].Desc, tc.seed[i].Created)
			if err != nil {
				t.Error(err)
			}
			testDb.imageIO.Put(tc.seed[i].Data, tc.seed[i].Ident)
		}
		got := testDb.GetAllImages()
		assert.Equal(t, tc.seed, got)
	}

}

func TestAllDocuments(t *testing.T) {
	testDb, db := newTestDb(t.TempDir(), true)

	type testcase struct {
		seed []Document
	}
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
		got := testDb.AllDocuments()
		assert.Equal(t, tc.seed, got)
	}

}

func TestUpdateDocument(t *testing.T) {
	type testcase struct {
		migrate bool
		seed    Document
		input   Document
		err     error
	}

	for _, tc := range []testcase{
		{
			migrate: true,
			seed: Document{
				Row:      1,
				Ident:    Identifier("qwerty"),
				Title:    "abc 123",
				Created:  "2024-12-31",
				Body:     "blog post body etc",
				Category: BLOG,
				Sample:   "this is a sample",
			},
			input: Document{
				Row:      1,
				Ident:    Identifier("qwerty"),
				Title:    "new title",
				Created:  "2024-12-31",
				Body:     "new updated post that must be reflected after the update",
				Category: BLOG,
				Sample:   "new updated post that must be reflected after the update",
			},
			err: nil,
		},
		{
			migrate: true,
			seed: Document{
				Row:      1,
				Ident:    Identifier("asdf"),
				Title:    "abc 123",
				Created:  "2024-12-31",
				Body:     "blog post body etc",
				Category: BLOG,
				Sample:   "this is a sample",
			},
			input: Document{
				Row:      1,
				Ident:    Identifier("This id does not exist"),
				Title:    "new title",
				Created:  "2024-12-31",
				Body:     "new updated post that must be reflected after the update",
				Category: BLOG,
				Sample:   "new updated post that must be reflected after the update",
			},
			err: ErrNotExists,
		},
		{
			migrate: false, // not creating the database tables so we can error out the SQL statement execution
			seed: Document{
				Row:      1,
				Ident:    Identifier("asdf"),
				Title:    "abc 123",
				Created:  "2024-12-31",
				Body:     "blog post body etc",
				Category: BLOG,
				Sample:   "this is a sample",
			},
			input: Document{
				Row:      1,
				Ident:    Identifier("This id does not exist"),
				Title:    "new title",
				Created:  "2024-12-31",
				Body:     "new updated post that must be reflected after the update",
				Category: BLOG,
				Sample:   "new updated post that must be reflected after the update",
			},
			err: errors.New("no such column: title"),
		},
	} {
		testDb, db := newTestDb(t.TempDir(), tc.migrate)
		if tc.migrate {
			stmt, _ := db.Prepare("INSERT INTO posts(id, title, created, body, category, sample) VALUES (?,?,?,?,?,?)")
			_, err := stmt.Exec(tc.seed.Ident, tc.seed.Title, tc.seed.Created, tc.seed.Body, tc.seed.Category, tc.seed.Sample)
			if err != nil {
				t.Error(err)
			}
		}
		err := testDb.UpdateDocument(tc.input)
		if err != nil {
			assert.Equal(t, tc.err.Error(), err.Error())
		} else {

			row := db.QueryRow("SELECT * FROM posts WHERE id = ?", tc.seed.Ident)
			var got Document
			if err := row.Scan(&got.Row, &got.Ident, &got.Title, &got.Created, &got.Body, &got.Category, &got.Sample); err != nil {
				assert.Equal(t, tc.err, err)
			}
			assert.Equal(t, tc.input, got)
		}

	}
}

func TestAddImage(t *testing.T) {

	type testcase struct {
		data  []byte
		title string
		desc  string
		err   error
	}
	testDb, _ := newTestDb(t.TempDir(), true)
	for _, tc := range []testcase{
		{
			data:  []byte("abc123xyz098"),
			title: "dont matter",
			desc:  "also dont matter",
		},
	} {
		id, err := testDb.AddImage(tc.data, tc.title, tc.desc)
		if err != nil {
			assert.Equal(t, tc.err, err)
		} else {
			b, err := testDb.imageIO.Get(id)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, tc.data, b)
		}

	}

}
func TestAddMenuItem(t *testing.T) {

	type testcase struct {
		input []LinkPair
		err   error
	}
	testDb, db := newTestDb(t.TempDir(), true)
	for _, tc := range []testcase{
		{
			input: []LinkPair{
				{
					Text: "abc 123",
					Link: "/abc/123",
				},
			},
			err: nil,
		},
	} {
		for i := range tc.input {
			err := testDb.AddMenuItem(tc.input[i])
			if err != nil {
				assert.Equal(t, tc.err, err)
			}
			rows, err := db.Query("SELECT * FROM menu")
			var got []LinkPair
			defer rows.Close()
			for rows.Next() {
				var id int
				var item LinkPair
				err = rows.Scan(&id, &item.Link, &item.Text)
				if err != nil {
					log.Fatal(err)
				}
				got = append(got, item)
			}
			assert.Equal(t, tc.input, got)
		}
	}
}
func TestAddNavbarItem(t *testing.T) {
	type testcase struct {
		input []NavBarItem
		err   error
	}
	testDb, db := newTestDb(t.TempDir(), true)
	for _, tc := range []testcase{
		{
			input: []NavBarItem{
				{
					Redirect: "",
					Link:     "",
					Png:      []byte(""),
				},
			},
		},
	} {
		for i := range tc.input {
			err := testDb.AddNavbarItem(tc.input[i])
			if err != nil {
				assert.Equal(t, tc.err, err)
			}

		}
		rows, err := db.Query("SELECT * FROM navbar")
		var got []NavBarItem
		defer rows.Close()
		for rows.Next() {
			var item NavBarItem
			var id int
			err = rows.Scan(&id, &item.Png, &item.Link, &item.Redirect)
			if err != nil {
				log.Fatal(err)
			}
			got = append(got, item)
		}
		assert.Equal(t, tc.input, got)

	}
}
func TestAddAsset(t *testing.T) {
	type testcase struct {
		input []Asset
		err   error
	}
	testDb, db := newTestDb(t.TempDir(), true)
	for _, tc := range []testcase{
		{
			input: []Asset{
				{
					Data: []byte(""),
					Name: "",
				},
			},
		},
	} {
		for i := range tc.input {
			err := testDb.AddAsset(tc.input[i].Name, tc.input[i].Data)
			if err != nil {
				assert.Equal(t, tc.err, err)
			}

		}
		rows, err := db.Query("SELECT * FROM assets")
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
	}
}
func TestAddDocument(t *testing.T) {

	testDb, db := newTestDb(t.TempDir(), true)
	type testcase struct {
		seed Document
		err  error
	}
	for _, tc := range []testcase{
		{
			seed: Document{
				Title:    "abc 123",
				Body:     "blog post body etc",
				Created:  "2024-12-31",
				Category: BLOG,
				Sample:   "this is a sample",
			},
			err: nil,
		},
	} {
		id, err := testDb.AddDocument(tc.seed)
		if err != nil {
			assert.Equal(t, tc.err, err)
		}
		row := db.QueryRow("SELECT * FROM posts WHERE id = ?", id)
		var got Document
		var rowNum int
		if err := row.Scan(&rowNum, &got.Ident, &got.Title, &got.Created, &got.Body, &got.Category, &got.Sample); err != nil {
			assert.Equal(t, tc.err, err)
		}
		want := Document{
			Ident:    id,
			Title:    tc.seed.Title,
			Body:     tc.seed.Body,
			Category: tc.seed.Category,
			Created:  tc.seed.Created,
			Sample:   tc.seed.MakeSample(),
		}

		assert.Equal(t, want, got)
	}
}
func TestAddAdminTableEntry(t *testing.T) {
	type testcase struct {
		input AdminPage
		err   error
	}
	testDb, db := newTestDb(t.TempDir(), true)
	for _, tc := range []testcase{
		{
			input: AdminPage{
				Tables: map[string][]TableData{
					"test category": {
						{
							DisplayName: "abc 123",
							Link:        "/abc/123",
						},
					},
				},
			},
			err: nil,
		},
	} {
		for ctg, tables := range tc.input.Tables {
			for i := range tables {
				err := testDb.AddAdminTableEntry(tables[i], ctg)
				if err != nil {
					assert.Equal(t, tc.err, err)
				}
			}
		}
		rows, err := db.Query("SELECT * FROM admin")
		got := AdminPage{Tables: map[string][]TableData{}}
		defer rows.Close()
		for rows.Next() {
			var item TableData
			var id int
			var category string
			err = rows.Scan(&id, &item.DisplayName, &item.Link, &category)
			if err != nil {
				log.Fatal(err)
			}
			got.Tables[category] = append(got.Tables[category], item)

		}
		assert.Equal(t, tc.input, got)
	}
}

func TestDeleteDocument(t *testing.T) {
	type testcase struct {
		input Document
		err   error
	}
	testDb, db := newTestDb(t.TempDir(), true)
	for _, tc := range []testcase{
		{
			input: Document{
				Title:    "abc 123",
				Body:     "blog post body etc",
				Created:  "2024-12-31",
				Category: BLOG,
				Sample:   "this is a sample",
			},
			err: nil,
		},
	} {
		id, err := testDb.AddDocument(tc.input)
		if err != nil {
			t.Error("failed to add doc: ", err)
		}
		err = testDb.DeleteDocument(id)
		if err != nil {
			assert.Equal(t, tc.err, err)
		}
		row, _ := db.Query("SELECT * FROM posts")
		if row.Next() {
			t.Error("Too many rows returned after deleting")
		}

	}
}

func TestGetImageStore(t *testing.T) {

	// testDb, db := newTestDb(t.TempDir(), true)
}
func TestNewIdentifier(t *testing.T) {

	// testDb, db := newTestDb(t.TempDir(), true)
}
