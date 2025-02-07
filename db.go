package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	sq "github.com/Masterminds/squirrel"
	_ "modernc.org/sqlite"
)

type db struct {
	conn    *sql.DB
	nextRow int
	columns int
}

func newDB() (*db, error) {
	conn, err := sql.Open("sqlite", ":memory:")
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
	_, err = conn.Exec("CREATE VIRTUAL TABLE ft USING fts5(value)")
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
	q = sq.Insert("ft").Columns("value")
	for _, value := range row {
		q = q.Values(value)
	}
	if _, err := q.RunWith(db.conn).Exec(); err != nil {
		return err
	}
	db.columns = max(db.columns, len(row))
	db.nextRow++
	return nil
}

func (db *db) getCell(row, column int) (string, error) {
	var sql strings.Builder
	sql.WriteString(fmt.Sprintf("SELECT c%d.value", column))
	sql.WriteString(" FROM (")
	for i := range db.columns {
		if i > 0 {
			sql.WriteString(" LEFT JOIN")
		}
		sql.WriteString("(")
		sql.WriteString("SELECT row, value FROM cells")
		sql.WriteString(fmt.Sprintf(" WHERE column = %d", i))
		sql.WriteString(")")
		sql.WriteString(fmt.Sprintf("AS c%d", i))
		if i > 0 {
			sql.WriteString(" USING (row)")
		}
	}
	sql.WriteString(" )")

	sql.WriteString(" ORDER BY ")
	for i := range db.columns {
		if i > 0 {
			sql.WriteString(", ")
		}
		sql.WriteString(fmt.Sprintf("c%d.value", i))
	}

	sql.WriteString(" LIMIT 1")
	sql.WriteString(fmt.Sprintf(" OFFSET %d", row))

	var value string
	if err := db.conn.QueryRow(sql.String()).Scan(&value); err != nil {
		return "", fmt.Errorf("%w: %s", err, sql.String())
	}

	return value, nil
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
