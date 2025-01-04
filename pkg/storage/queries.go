package storage

const postsTable = `
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
const imagesTable = `
	CREATE TABLE IF NOT EXISTS images(
		row INTEGER PRIMARY KEY AUTOINCREMENT,
		id TEXT NOT NULL,
		title TEXT NOT NULL,
		desc TEXT NOT NULL,
		created TEXT NOT NULL
	);
	`
const menuItemsTable = `
	CREATE TABLE IF NOT EXISTS menu(
		row INTEGER PRIMARY KEY AUTOINCREMENT,
		link TEXT NOT NULL,
		text TEXT NOT NULL
	);
	`
const navbarItemsTable = `
	CREATE TABLE IF NOT EXISTS navbar(
		row INTEGER PRIMARY KEY AUTOINCREMENT,
		png BLOB NOT NULL,
		link TEXT NOT NULL,
		redirect TEXT
	);`
const assetTable = `
	CREATE TABLE IF NOT EXISTS assets(
		row INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		data BLOB NOT NULL
	);
	`
const adminTable = `
	CREATE TABLE IF NOT EXISTS admin(
		row INTEGER PRIMARY KEY AUTOINCREMENT,
		display_name TEXT NOT NULL,
		link TEXT NOT NULL,
		category TEXT NOT NULL
	);
	`

var RequiredTables = []string{postsTable, imagesTable, menuItemsTable, navbarItemsTable, assetTable, adminTable}
