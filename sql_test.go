package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQL_getWindowSQL(t *testing.T) {
	tests := []struct {
		name    string
		columns int
		size    int
		rowid   *int
		filter  string
		order   []SortOrder
		values  []string
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
			name:    "window with rowid but no order and no filter",
			columns: 2,
			size:    3,
			rowid:   intPtr(5),
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
					WHERE row <= 5
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
					WHERE row >= 5
					ORDER BY row
					LIMIT 3
				)
			`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := getWindowSQL(
				tt.columns,
				tt.size,
				tt.rowid,
				tt.order,
				tt.filter,
				tt.values...,
			)
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

func intPtr(i int) *int {
	return &i
}
