package main

import (
	"fmt"
	"strconv"
	"strings"
)

// order by c0.value, c1.value, row
// -->
// where c0.value > "marx" or (c0.value = "marx" and c1.value > "german
// ideology") or row = 4
//
// order by c0.value desc, c1.value, row
// -->
// where c0.value < "marx" or (c0.value = "marx" and c1.value > "german
// ideology") or row = 4
//
// order by row
// -->
// where row >= 4

// getWindowSQL builds the SQL string and args for retrieving the cursor window from the DB.
func getWindowSQL(columns, size int, rowid *int, order []SortOrder, filter string, values ...string) (string, []any) {
	joinCellsSQL, joinCellsSQLArgs := getJoinCellsSQL(columns, size, rowid, order, false, filter, values...)
	if rowid == nil {
		return joinCellsSQL, joinCellsSQLArgs
	}

	var b strings.Builder
	b.WriteString(joinCellsSQL)
	b.WriteString(" UNION ALL ")

	joinCellsBeforeSQL, joinCellsBeforeArgs := getJoinCellsSQL(columns, size, rowid, order, true, filter, values...)
	b.WriteString(joinCellsBeforeSQL)

	return b.String(), append(joinCellsSQLArgs, joinCellsBeforeArgs...)
}

func getJoinCellsSQL(columns, size int, rowid *int, order []SortOrder, reverse bool, filter string, values ...string) (string, []any) {
	var (
		joinCells strings.Builder
		args      []any
	)
	joinCells.WriteString("SELECT *")
	joinCells.WriteString(" FROM ( ")
	for i := range columns {
		if i > 0 {
			joinCells.WriteString(" LEFT JOIN")
		}
		joinCells.WriteString("(")
		joinCells.WriteString(" SELECT row, value FROM cells")
		joinCells.WriteString(fmt.Sprintf(" WHERE column = %d", i))
		joinCells.WriteString(" )")
		joinCells.WriteString(fmt.Sprintf(" AS c%d", i))
		if i > 0 {
			joinCells.WriteString(" USING (row)")
		}
	}
	joinCells.WriteString(" )")

	// If rowid is nil then there is no where condition and only the first n rows
	// are queried.
	if rowid == nil {
		joinCells.WriteString("ORDER BY row")
		joinCells.WriteString("LIMIT ")
		joinCells.WriteString(strconv.Itoa(size))
		return joinCells.String(), nil
	}

	joinCells.WriteString("WHERE ")

	if filter != "" {
		for i := range columns {
			joinCells.WriteString(fmt.Sprintf("c%d.value", i))
			joinCells.WriteString(" LIKE ? ")
			args = append(args, "%"+filter+"%")
		}
	}

	if len(order) == 0 {
		joinCells.WriteString(" row")
		if reverse {
			joinCells.WriteString("<= ")
		} else {
			joinCells.WriteString(">= ")
		}
		joinCells.WriteString(strconv.Itoa(*rowid))
		joinCells.WriteString("ORDER BY row")
	} else {
		for i, s := range order {
			joinCells.WriteString(fmt.Sprintf("c%d.value ", s.Column))
			if reverse {
				if s.Descending {
					joinCells.WriteString("> ")
				} else {
					joinCells.WriteString("< ")
				}
			} else {
				if s.Descending {
					joinCells.WriteString("< ")
				} else {
					joinCells.WriteString("> ")
				}
			}
			joinCells.WriteString(values[i])
		}
		joinCells.WriteString(" OR row = ")
		joinCells.WriteString(strconv.Itoa(*rowid))

		joinCells.WriteString(" ORDER BY")
		for i, s := range order {
			if i > 0 {
				joinCells.WriteString(", ")
			}
			joinCells.WriteString(fmt.Sprintf("c%d.value ", s.Column))
			if s.Descending {
				joinCells.WriteString("< ")
			} else {
				joinCells.WriteString("> ")
			}
			joinCells.WriteString(values[i])
		}
		joinCells.WriteString(", row")
	}

	joinCells.WriteString(" LIMIT ")
	joinCells.WriteString(strconv.Itoa(size))

	return joinCells.String(), args
}
