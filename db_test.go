package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB_moveCursorSQL(t *testing.T) {
	tests := []struct {
		name    string
		columns int
		cursor  string
		size    int
		want    string
	}{
		{
			name:    "init",
			columns: 2,
			size:    3,
			want: `
				SELECT *
				FROM (
					(
						SELECT row, value FROM cells WHERE column = 0
					) AS c0
					LEFT JOIN(
						SELECT row, value FROM cells WHERE column = 1
					) AS c1 USING (row)
				)
				ORDER BY row
				LIMIT 3
			`,
		},
		{
			name:    "move cursor down one row",
			columns: 2,
			size:    3,
			cursor:  ` WHERE rowid >= 23`,
			want: `
				SELECT * FROM (
					SELECT *
					FROM (
						(
							SELECT row, value FROM cells WHERE column = 0
						) AS c0
						LEFT JOIN(
							SELECT row, value FROM cells WHERE column = 1
						) AS c1 USING (row)
					)
					WHERE row <= 1
					ORDER BY row DESC
					LIMIT 3
				)
				UNION ALL
				SELECT * FROM (
					SELECT *
					FROM (
						(
							SELECT row, value FROM cells WHERE column = 0
						) AS c0
						LEFT JOIN(
							SELECT row, value FROM cells WHERE column = 1
						) AS c1 USING (row)
					)
					WHERE row >= 3
					ORDER BY row
					LIMIT 3
				)
			`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := newDB()
			require.NoError(t, err)
			db.columns = tt.columns
			db.cursor = tt.cursor
			sql, args := db.moveCursorSQL(0, moveCursorOptions{size: tt.size})
			assert.Equal(t, trimSQL(tt.want), sql)
			assert.Len(t, args, 0)
		})
	}
}

func trimSQL(sql string) string {
	sql = strings.TrimSpace(sql)
	sql = strings.Join(strings.Fields(sql), " ")
	return sql
}
