package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestData(t *testing.T) {
	tests := []struct {
		name string
		do   func(*data)
		want func(t *testing.T, data *data)
	}{
		{
			name: "retrieve cells",
			do: func(d *data) {
				d.Append(
					[]string{"Marx & Engels", "The German Ideology"},
					[]string{"Thomas Mann", "Death in Venice and Other Stories"},
				)
			},
			want: func(t *testing.T, d *data) {
				assert.Equal(t, "Marx & Engels", d.At(0, 0))
				assert.Equal(t, "The German Ideology", d.At(0, 1))
				assert.Equal(t, "Thomas Mann", d.At(1, 0))
				assert.Equal(t, "Death in Venice and Other Stories", d.At(1, 1))
			},
		},
		{
			name: "window restricts visible rows",
			do: func(d *data) {
				d.window.start = 1
				d.Append(
					[]string{"Marx & Engels", "The German Ideology"},
					[]string{"Thomas Mann", "Death in Venice and Other Stories"},
					[]string{"Charles Bukowski", "Ham and Rye"},
				)
			},
			want: func(t *testing.T, d *data) {
				assert.Equal(t, "Thomas Mann", d.At(0, 0))
				assert.Equal(t, "Death in Venice and Other Stories", d.At(0, 1))
				assert.Equal(t, "Charles Bukowski", d.At(1, 0))
				assert.Equal(t, "Ham and Rye", d.At(1, 1))
			},
		},
		{
			name: "move cursor down one row",
			do: func(d *data) {
				d.Append(
					[]string{"Marx & Engels", "The German Ideology"},
					[]string{"Thomas Mann", "Death in Venice and Other Stories"},
				)
				d.moveCursor(1)
			},
			want: func(t *testing.T, d *data) {
				assert.Equal(t, d.cursor, 1)
			},
		},
		{
			name: "move cursor down and up one row",
			do: func(d *data) {
				d.Append(
					[]string{"Marx & Engels", "The German Ideology"},
					[]string{"Thomas Mann", "Death in Venice and Other Stories"},
				)
				d.moveCursor(1)
				d.moveCursor(-1)
			},
			want: func(t *testing.T, d *data) {
				assert.Equal(t, d.cursor, 0)
			},
		},
		{
			name: "cursor cannot move beyond last row",
			do: func(d *data) {
				d.Append(
					[]string{"Marx & Engels", "The German Ideology"},
					[]string{"Thomas Mann", "Death in Venice and Other Stories"},
				)
				d.moveCursor(99)
			},
			want: func(t *testing.T, d *data) {
				assert.Equal(t, d.cursor, 1)
			},
		},
		{
			name: "cursor cannot move before first row",
			do: func(d *data) {
				d.Append(
					[]string{"Marx & Engels", "The German Ideology"},
					[]string{"Thomas Mann", "Death in Venice and Other Stories"},
				)
				d.moveCursor(-99)
			},
			want: func(t *testing.T, d *data) {
				assert.Equal(t, d.cursor, 0)
			},
		},
		{
			name: "moving window beyond cursor moves cursor",
			do: func(d *data) {
				d.window.size = 1
				d.Append(
					[]string{"Marx & Engels", "The German Ideology"},
					[]string{"Thomas Mann", "Death in Venice and Other Stories"},
				)
				d.moveWindow(1)
			},
			want: func(t *testing.T, d *data) {
				assert.Equal(t, d.window.start, 1)
				assert.Equal(t, d.cursor, 1)
			},
		},
		{
			name: "move window but cursor does not need moving",
			do: func(d *data) {
				d.window.size = 1
				d.Append(
					[]string{"Marx & Engels", "The German Ideology"},
					[]string{"Thomas Mann", "Death in Venice and Other Stories"},
				)
				d.moveCursor(1)
				d.moveWindow(1)
			},
			want: func(t *testing.T, d *data) {
				assert.Equal(t, d.window.start, 1)
				assert.Equal(t, d.cursor, 1)
			},
		},
		{
			name: "filter rows",
			do: func(d *data) {
				d.window.size = 5
				d.filter = "Ernest"
				d.Append(
					[]string{"Marx & Engels", "The German Ideology"},
					[]string{"Ernest Hemingway", "To Have and Have Not"},
					[]string{"Ernest Hemingway", "The Sun Also Rises"},
					[]string{"Thomas Mann", "Death in Venice and Other Stories"},
					[]string{"Ernest Hemingway", "A Farewell to Arms"},
				)
			},
			want: func(t *testing.T, d *data) {
				assert.Equal(t, 3, d.Rows())
				assert.Equal(t, "Ernest Hemingway", d.At(0, 0))
				assert.Equal(t, "To Have and Have Not", d.At(0, 1))
				assert.Equal(t, "Ernest Hemingway", d.At(1, 0))
				assert.Equal(t, "The Sun Also Rises", d.At(1, 1))
				assert.Equal(t, "Ernest Hemingway", d.At(2, 0))
				assert.Equal(t, "A Farewell to Arms", d.At(2, 1))
			},
		},
		{
			name: "sort rows",
			do: func(d *data) {
				d.window.size = 5
				d.sort = []SortOrder{{Column: 0, Descending: true}, {Column: 1}}
				d.Append(
					[]string{"Marx & Engels", "The German Ideology"},
					[]string{"Ernest Hemingway", "To Have and Have Not"},
					[]string{"Ernest Hemingway", "The Sun Also Rises"},
					[]string{"Thomas Mann", "Death in Venice and Other Stories"},
					[]string{"Ernest Hemingway", "A Farewell to Arms"},
				)
			},
			want: func(t *testing.T, d *data) {
				assert.Equal(t, "Thomas Mann", d.At(0, 0))
				assert.Equal(t, "Death in Venice and Other Stories", d.At(0, 1))
				assert.Equal(t, "Marx & Engels", d.At(1, 0))
				assert.Equal(t, "The German Ideology", d.At(1, 1))
				assert.Equal(t, "Ernest Hemingway", d.At(2, 0))
				assert.Equal(t, "A Farewell to Arms", d.At(2, 1))
				assert.Equal(t, "Ernest Hemingway", d.At(3, 0))
				assert.Equal(t, "The Sun Also Rises", d.At(3, 1))
				assert.Equal(t, "Ernest Hemingway", d.At(4, 0))
				assert.Equal(t, "To Have and Have Not", d.At(4, 1))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := newData()
			require.NoError(t, err)
			tt.do(data)
			tt.want(t, data)
		})
	}
}
