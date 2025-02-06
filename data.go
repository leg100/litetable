package main

type data struct {
	db     *db
	window window
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
	return d.db.getCell(d.window.start+row, column)
}

func (d *data) Columns() int { return d.db.columns }

func (d *data) Rows() int {
	return min(d.db.getRowCount()-d.window.start, d.window.size)
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
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
