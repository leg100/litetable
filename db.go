package main

import (
	"database/sql"
	"log"

	sq "github.com/Masterminds/squirrel"
	_ "modernc.org/sqlite"
)

type db struct {
	MaxColumns int
	conn       *sql.DB
	nextRow    int
}

func newDB() (*db, error) {
	conn, err := sql.Open("sqlite", "file::memory:")
	if err != nil {
		return nil, err
	}
	_, err = conn.Exec("CREATE TABLE cells(row, column, value, PRIMARY KEY(row, column))")
	if err != nil {
		return nil, err
	}
	_, err = conn.Exec("CREATE INDEX indexed_value ON cells (value)")
	if err != nil {
		return nil, err
	}
	return &db{conn: conn}, nil
}

func (db *db) insertRow(row []string) error {
	q := sq.Insert("cells").Columns("row", "column", "value")
	for i, value := range row {
		q = q.Values(db.nextRow, i, value)
	}
	if _, err := q.RunWith(db.conn).Exec(); err != nil {
		return err
	}
	db.MaxColumns = max(db.MaxColumns, len(row))
	db.nextRow++
	return nil
}

func (db *db) getCell(row, column int) string {
	var value string
	err := db.conn.QueryRow("SELECT value FROM cells WHERE column = ? ORDER BY value LIMIT 1 OFFSET ?", column, row).Scan(&value)
	if err != nil {
		log.Fatal("At(): ", err.Error())
	}
	return value
}

func (db *db) getRowCount() int {
	var rows int
	err := db.conn.QueryRow("SELECT count(distinct(row)) FROM cells").Scan(&rows)
	if err != nil {
		log.Fatal("Rows(): ", err.Error())
		return 0
	}
	return rows
}
