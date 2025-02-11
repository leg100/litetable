package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
	_ "modernc.org/sqlite"
)

type db struct {
	conn    *sql.DB
	nextRow int
	columns int
	cursor  string
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

type window struct {
	cells  [][]string
	cursor int
	start  int
}

type moveCursorOptions struct {
	size   int
	filter string
	sort   []SortOrder
}

func (db *db) moveCursor(n int, opts moveCursorOptions) (window, error) {
	sqls, args := getJoinCellsSQL(0, 0, nil, nil, "")

	var (
		rowid  int
		values = make([]string, db.columns)
		dest   = make([]any, 1+len(values))
	)
	dest[0] = &rowid
	for i := range db.columns {
		dest[i+1] = &values[i]
	}
	result, err := db.conn.Query(sqls, args...)
	if errors.Is(err, sql.ErrNoRows) {
		// Cursor is already on first/last row and cannot be moved.
		return window{}, nil
	} else if err != nil {
		return window{}, fmt.Errorf("%w: %s", err, sqls)
	}
	defer result.Close()

	for result.Next() {
		if err := result.Scan(dest); err != nil {
			return window{}, err
		}
	}
	if err := result.Err(); err != nil {
		return window{}, err
	}

	// SQL query succeeded; now compose a query that'll be used to retrieve the
	// cursor in future.
	if len(opts.sort) == 0 {
		db.cursor = " WHERE rowid >= " + strconv.Itoa(rowid)
	} else {
		var cursor strings.Builder
		cursor.WriteString(" WHERE ")
		for i, order := range opts.sort {
			if i > 0 {
				cursor.WriteString(" AND ")
			}
			cursor.WriteString(fmt.Sprintf("c%d.value ", order.Column))
			if order.Descending {
				cursor.WriteString("<= ")
			} else {
				cursor.WriteString(">= ")
			}
			cursor.WriteString(values[order.Column])
		}
		cursor.WriteString(" AND rowid = ")
		cursor.WriteString(strconv.Itoa(rowid))
		db.cursor = cursor.String()
	}

	return window{}, nil
}

type getCellOptions struct {
	filter string
	sort   []SortOrder
}

func (db *db) getCell(row, column int, opts getCellOptions) (string, error) {
	var (
		b    strings.Builder
		args []any
	)
	b.WriteString(fmt.Sprintf("SELECT c%d.value", column))
	b.WriteString(" FROM (")
	for i := range db.columns {
		if i > 0 {
			b.WriteString(" LEFT JOIN")
		}
		b.WriteString("(")
		b.WriteString("SELECT row, value FROM cells")
		b.WriteString(fmt.Sprintf(" WHERE column = %d", i))
		b.WriteString(")")
		b.WriteString(fmt.Sprintf(" AS c%d", i))
		if i > 0 {
			b.WriteString(" USING (row)")
		}
	}
	b.WriteString(" )")

	if opts.filter != "" {
		for i := range db.columns {
			if i > 0 {
				b.WriteString(" OR")
			} else {
				b.WriteString(" WHERE")
			}
			b.WriteString(fmt.Sprintf(" c%d.value LIKE ?", i))
			args = append(args, "%"+opts.filter+"%")
		}
	}

	if len(opts.sort) > 0 {
		b.WriteString(" ORDER BY ")
		for i, order := range opts.sort {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("c%d.value", order.Column))
			if order.Descending {
				b.WriteString(" DESC")
			}
		}
	}

	b.WriteString(" LIMIT 1")
	b.WriteString(fmt.Sprintf(" OFFSET %d", row))

	var value string
	if err := db.conn.QueryRow(b.String(), args...).Scan(&value); err != nil {
		return "", fmt.Errorf("%w: %s", err, b.String())
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
