package tapirdb

import (
	"context"
	"errors"
	"fmt"

	tapir "github.com/ViolaChenYT/TAPIR"
	"github.com/magiconair/properties"
	"github.com/pingcap/go-ycsb/pkg/ycsb"
)

type TapirDB struct {
	app *tapir.TapirApp
}

type TapirCreator struct{}

func (c TapirCreator) Create(p *properties.Properties) (ycsb.DB, error) {
	d := new(TapirDB)
	return d, nil
}

// CreateTapirDB creates a new instance of the TapirDB.
func CreateTapirDB() *TapirDB {
	return &TapirDB{
		app: tapir.NewTapirApp(),
	}
}

// Close closes the database layer.
func (d *TapirDB) Close() error {
	fmt.Println("Closing the database layer")
	return nil
}

// InitThread initializes the state associated with the goroutine worker.
func (d *TapirDB) InitThread(ctx context.Context, threadID int, threadCount int) context.Context {
	fmt.Printf("Initializing thread %d out of %d\n", threadID, threadCount)
	return ctx
}

// CleanupThread cleans up the state when the worker finishes.
func (d *TapirDB) CleanupThread(ctx context.Context) {
	fmt.Println("Cleaning up thread")
}

// Read reads a record from the database and returns a map of each field/value pair.
func (d *TapirDB) Read(ctx context.Context, table string, key string, fields []string) (map[string][]byte, error) {
	fmt.Printf("Reading record with key %s from table %s\n", key, table)
	return d.app.Read(table, key, fields)
}

// Scan scans records from the database.
func (d *TapirDB) Scan(ctx context.Context, table string, startKey string, count int, fields []string) ([]map[string][]byte, error) {
	return nil, errors.New("Scan operation not implemented!")
}

// Update updates a record in the database.
func (d *TapirDB) Update(ctx context.Context, table string, key string, values map[string][]byte) error {
	fmt.Printf("Updating record with key %s in table %s\n", key, table)
	return d.app.Update(table, key, values)
}

// Insert inserts a record into the database.
func (d *TapirDB) Insert(ctx context.Context, table string, key string, values map[string][]byte) error {
	fmt.Printf("Inserting record with key %s into table %s\n", key, table)
	return d.app.Insert(table, key, values)
}

// Delete deletes a record from the database.
func (d *TapirDB) Delete(ctx context.Context, table string, key string) error {
	return d.app.Delete(table, key)
}

func (d *TapirDB) Start() error {
	fmt.Printf("Starting a transaction\n")
	return d.app.Start()
}

func (d *TapirDB) Commit() error {
	fmt.Printf("Committing a transaction\n")
	return d.app.Commit()
}

func (d *TapirDB) Abort() error {
	fmt.Printf("Aborting a transaction\n")
	return d.app.Abort()
}

// Register with the server
func init() {
	ycsb.RegisterDBCreator("tapir", TapirCreator{})
}
