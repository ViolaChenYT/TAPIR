package tapir

import (
	"bufio" //
	"fmt"
	"net"
	"net/rpc"
	"time"
)

type TapirClient struct {
	int client_id;

	// Transport for IR clients
}

func NewClient(id int, closestReplica int) (*TapirClient, error) {
	client := TapirClient{
		client_id:        id
	}

	// Run the transport in a new thread
	go client.run_client()

	return &client, nil
}

// Runs the transport event loop.
func (c *TapirClient) run_client() {
	// TODO
}

func (c *TapirClient) Prepare(timestamp Time.time) {
	
}

func (c *TapirClient) Begin() {
	// Implementation for beginning a transaction
}

// Get the value corresponding to key.
func (c *TapirClient) Get(key string) (string, error) {
	// Implementation for getting a value
}

// Set the value for the given key.
func (c *TapirClient) Put(key string, value string) error {
	// Implementation for setting a value
}

// Commit all Get(s) and Put(s) since Begin().
func (c *TapirClient) Commit() bool {
	// Implementation for committing a transaction
}

// Abort all Get(s) and Put(s) since Begin().
func (c *TapirClient) Abort() {
	// Implementation for aborting a transaction
}
