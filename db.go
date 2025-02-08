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
	_, err = conn.Exec("CREATE INDEX idx_value ON cells (value)")
	if err != nil {
		return nil, err
	}
	_, err = conn.Exec("CREATE TABLE cursor(row, PRIMARY KEY(row))")
	if err != nil {
		return nil, err
	}
	return &db{conn: conn}, nil
}

func (db *db) insertRow(rows ...[]string) error {
	q := sq.Insert("cells").Columns("row", "column", "value")
	for _, row := range rows {
		for col, value := range row {
			q = q.Values(db.nextRow, col, value)
		}
		// TODO: only do this when there is no error
		db.nextRow++
		db.columns = max(db.columns, len(row))
	}
	if _, err := q.RunWith(db.conn).Exec(); err != nil {
		return err
	}
	return nil
}

type moveCursorOptions struct {
	filter string
	sort   []SortOrder
}

func (db *db) moveCursor(n int, opts moveCursorOptions) error {
	var sql strings.Builder
	sql.WriteString("UPDATE cursor")
	sql.WriteString(" SET row = new_cursor")
	sql.WriteString(" FROM (")
	sql.WriteString(" SELECT * FROM (")
	sql.WriteString(" SELECT row, IFNULL(LEAD(row, ?) OVER w, LAST_VALUE(row) OVER w) AS new_cursor FROM (")
	for i := range db.columns {
		if i > 0 {
			sql.WriteString(" LEFT JOIN")
		}
		sql.WriteString("(")
		sql.WriteString("SELECT row, value FROM cells")
		sql.WriteString(fmt.Sprintf(" WHERE column = %d", i))
		sql.WriteString(")")
		sql.WriteString(fmt.Sprintf(" AS c%d", i))
		if i > 0 {
			sql.WriteString(" USING (row)")
		}
	}
	sql.WriteString(" )")

	if opts.filter != "" {
		for i := range db.columns {
			if i > 0 {
				sql.WriteString(" OR")
			} else {
				sql.WriteString(" WHERE")
			}
			sql.WriteString(fmt.Sprintf(" c%d.value LIKE '%%%s%%'", i, opts.filter))
		}
	}

	sql.WriteString("WINDOW w AS (")

	if len(opts.sort) > 0 {
		sql.WriteString(" ORDER BY ")
		for i, order := range opts.sort {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(fmt.Sprintf("c%d.value", order.Column))
			if order.Descending {
				sql.WriteString(" DESC")
			}
		}
	}
	// sqlite> update cursor set row = new_cursor from (select * from (select row, ifnull(lead(row, 3) over w, last_value(row) over w) as new_cursor from ((select row, value from cells where column = 0) as c0 left join (select row, value from cells where column = 1) as c1 using(row)) where c0.value like '%E%' window w as (order by c0.value rows between unbounded preceding and unbounded following))) as rows where cursor.row = rows.row;

	sql.WriteString("ROWS BETWEEN CURRENT ROW AND UNBOUNDED FOLLOWING)")

	sql.WriteString(") WHERE row = ")
	sql.WriteString(fmt.Sprintf(" OFFSET %d", row))

	var value string
	if err := db.conn.QueryRow(sql.String()).Scan(&value); err != nil {
		return "", fmt.Errorf("%w: %s", err, sql.String())
	}

	return value, nil
}

type getCellOptions struct {
	filter string
	sort   []SortOrder
}

func (db *db) getCell(row, column int, opts getCellOptions) (string, error) {
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
		sql.WriteString(fmt.Sprintf(" AS c%d", i))
		if i > 0 {
			sql.WriteString(" USING (row)")
		}
	}
	sql.WriteString(" )")

	if opts.filter != "" {
		for i := range db.columns {
			if i > 0 {
				sql.WriteString(" OR")
			} else {
				sql.WriteString(" WHERE")
			}
			sql.WriteString(fmt.Sprintf(" c%d.value LIKE '%%%s%%'", i, opts.filter))
		}
	}

	if len(opts.sort) > 0 {
		sql.WriteString(" ORDER BY ")
		for i, order := range opts.sort {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(fmt.Sprintf("c%d.value", order.Column))
			if order.Descending {
				sql.WriteString(" DESC")
			}
		}
	}

	sql.WriteString(" LIMIT 1")
	sql.WriteString(fmt.Sprintf(" OFFSET %d", row))

	var value string
	if err := db.conn.QueryRow(sql.String()).Scan(&value); err != nil {
		return "", fmt.Errorf("%w: %s", err, sql.String())
	}

	return value, nil
}

func (db *db) getRowCount(filter string) int {
	filter = "%" + filter + "%"
	var rows int
	err := db.conn.QueryRow("SELECT count(distinct(row)) FROM cells WHERE value LIKE ?", filter).Scan(&rows)
	if err != nil {
		log.Fatal("Rows(): ", err.Error())
		return 0
	}
	return rows
}
