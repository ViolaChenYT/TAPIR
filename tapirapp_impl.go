package tapir

import (
	"errors"
	"fmt"
	"strings"

	. "github.com/ViolaChenYT/TAPIR/tapir"
)

// TapirDB represents the implementation of the TapirApp interface.
type TapirAppImpl struct {
	client TapirClient
}

// NewTapirDB creates a new TapirDB instance.
func NewTapirDB() TapirApp {
	client, err := NewClient(0, 0)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return nil
	}
	return &TapirAppImpl{
		client: client,
	}
}

// Read reads a record from the database and returns a map of each field/value pair.
func (app *TapirAppImpl) Read(table string, key string, fields []string) (map[string][]byte, error) {
	val, err := app.client.Read(table + key)

	if err != nil || val == "" {
		return nil, err
	}

	row := NewTableRow(val)
	return row.FilterFields(fields)
}

// Update updates a record in the database.
func (app *TapirAppImpl) Update(table string, key string, values map[string][]byte) error {
	val, err := app.client.Read(table + key)
	var row TableRow = values

	if err != nil || val == "" {
		return errors.New("Key to update does not exist, key: " + table + key)
	}

	// Update values
	existingRow := NewTableRow(val)
	existingRow.Merge(row)
	return app.client.Write(table+key, existingRow.String())
}

// Insert inserts a record into the database.
func (app *TapirAppImpl) Insert(table string, key string, values map[string][]byte) error {
	val, err := app.client.Read(table + key)
	var row TableRow = values

	if err != nil || val == "" {
		// Key does not exist, insert the whole row
		app.client.Write(table+key, row.String())
	}

	// If key exists, merge with new values
	existingRow := NewTableRow(val)
	existingRow.Merge(row)
	return app.client.Write(table+key, existingRow.String())
}

// Delete deletes a record from the database.
func (app *TapirAppImpl) Delete(table string, key string) error {
	val, err := app.client.Read(table + key)

	if err != nil || val == "" {
		return errors.New("Key to be delete not exist, key: " + table + key)
	}
	// Zero out the record
	return app.client.Write(table+key, "")
}

// Start starts a transaction.
func (app *TapirAppImpl) Start() error {
	app.client.Begin()
	return nil
}

// Commit commits a transaction.
func (app *TapirAppImpl) Commit() error {
	ok := app.client.Commit()
	if ok {
		return nil
	} else {
		return errors.New("Commit failed.")
	}
}

// Abort aborts a transaction.
func (app *TapirAppImpl) Abort() error {
	app.client.Abort()
	return nil
}

/** Helper */

// TableRow is a map representing a table row with key-value pairs.
type TableRow map[string][]byte

// NewTableRowFromString creates a new TableRow instance from a string representation.
func NewTableRow(s string) TableRow {
	row := make(TableRow)
	columns := strings.Split(s, "\n")
	for _, c := range columns {
		column := strings.Split(c, "\t")
		if len(column) == 2 {
			row[column[0]] = []byte(column[1])
		}
	}
	return row
}

func (r TableRow) FilterFields(fields []string) (TableRow, error) {
	if fields == nil {
		return r, nil
	}
	subset := make(TableRow)
	for _, field := range fields {
		value, ok := r[field]
		if !ok {
			return nil, errors.New("Record does not contain field: " + field + "\nRecord: " + r.String())
		}
		subset[field] = value
	}
	return subset, nil
}

func (r TableRow) Merge(newRow TableRow) {
	for key, field := range newRow {
		r[key] = field
	}
}

// String returns the string representation of the TableRow.
func (r TableRow) String() string {
	var ret strings.Builder
	for key, value := range r {
		ret.WriteString(fmt.Sprintf("%s\t%s\n", key, string(value)))
	}
	return ret.String()
}
