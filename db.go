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
	db.columns = max(db.columns, len(row))
	db.nextRow++
	return nil
}

func (db *db) getCell(row, column int) string {
	var sql strings.Builder
	sql.WriteString("SELECT ")
	sql.WriteString(fmt.Sprintf("c%d.value", column))
	sql.WriteString(" FROM (")
	for i := range db.columns {
		if i > 0 {
			sql.WriteString(" LEFT JOIN ")
		}
		sql.WriteString("(SELECT row, value FROM cells ")
		sql.WriteString(fmt.Sprintf("WHERE column = %d", i))
		sql.WriteString(")")
		sql.WriteString(" AS ")
		sql.WriteString(fmt.Sprintf("c%d", i))
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
	sql.WriteString(" LIMIT 1 OFFSET ")
	sql.WriteString(fmt.Sprintf("%d", row))
	// log.Println(sql.String())
	var value string
	err := db.conn.QueryRow(sql.String()).Scan(&value)
	if err != nil {
		log.Fatal("getCell(): ", err.Error())
	}
	return value
	// sqlite> select c1.value as "0", c2.value as "1" from (select row, value from cells where column = 0) as c1 join (select row, value from cells where column = 1) as c2 using (row) order by 1, 2;
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
