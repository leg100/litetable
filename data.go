package main

import "log"

type data struct {
	db     *db
	window window
	cursor int
	filter string
	sort   []SortOrder
}

type window struct {
	start int
	size  int
}

type SortOrder struct {
	Column     int
	Descending bool
}

func newData() (*data, error) {
	db, err := newDB()
	if err != nil {
		return nil, err
	}
	return &data{
		db:     db,
		filter: "",
	}, nil
}

func (d *data) Append(rows ...[]string) error {
	return d.db.insertRow(rows...)
}

func (d *data) At(row, column int) string {
	v, err := d.db.getCell(d.window.start+row, column, getCellOptions{
		filter: d.filter,
		sort:   d.sort,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	return v
}

func (d *data) Columns() int { return d.db.columns }

func (d *data) Rows() int {
	return min(d.db.getRowCount(d.filter)-d.window.start, d.window.size)
}

func (d *data) PreviousRow() {
	d.moveCursor(-1)
}

func (d *data) NextRow() {
	d.moveCursor(1)
}

func (d *data) PageUp() {
	d.moveWindow(-d.window.size)
}

func (d *data) PageDown() {
	d.moveWindow(d.window.size)
}

func (d *data) moveCursor(n int) {
	d.cursor = clamp(d.cursor+n, 0, d.lastRowIndex())
	startMin := max(0, d.cursor-d.window.size+1)
	startMax := min(d.cursor, d.db.getRowCount(d.filter)-d.window.size)
	d.window.start = clamp(d.window.start, startMin, startMax)
}

func (d *data) moveWindow(rows int) {
	d.window.start = clamp(d.window.start+rows, 0, d.lastRowIndex())
	d.cursor = d.window.start
}

func (d *data) lastRowIndex() int {
	return max(0, d.db.getRowCount(d.filter)-1)
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
