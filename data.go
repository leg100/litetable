package main

import "log"

type data struct {
	db     *db
	window window
	cursor int
}

type window struct {
	start int
	size  int
}

func newData() (*data, error) {
	db, err := newDB()
	if err != nil {
		return nil, err
	}
	return &data{db: db}, nil
}

func (d *data) Append(row []string) error {
	return d.db.insertRow(row)
}

func (d *data) At(row, column int) string {
	if column == 0 {
		if (d.window.start + row) == d.cursor {
			return "✓"
		}
		return ""
	}
	column--
	v, err := d.db.getCell(d.window.start+row, column)
	if err != nil {
		log.Fatal(err.Error())
	}
	return v
}

func (d *data) Columns() int { return 1 + d.db.columns }

func (d *data) Rows() int {
	return min(d.db.getRowCount()-d.window.start, d.window.size)
}

func (d *data) PreviousRow() {
	d.moveCursor(-1)
}

func (d *data) NextRow() {
	d.moveCursor(1)
}

func (d *data) moveCursor(n int) {
	d.cursor = clamp(d.cursor+n, 0, max(0, d.db.getRowCount()-1))
	startMin := max(0, d.cursor-d.window.size+1)
	startMax := min(d.cursor, d.db.getRowCount()-d.window.size)
	d.window.start = clamp(d.window.start, startMin, startMax)
}

func (d *data) PageUp() {
	d.moveWindow(-d.window.size)
}

func (d *data) PageDown() {
	d.moveWindow(d.window.size)
}

func (d *data) moveWindow(rows int) {
	lastRowIndex := max(0, d.db.getRowCount()-1)
	d.window.start = clamp(d.window.start+rows, 0, lastRowIndex)
	d.cursor = d.window.start
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
