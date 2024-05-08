package IR

// test adding and initiating servers and replicas

// func TestRPCConnection(t *testing.T) {
// 	// Setup the server
// 	port := 9090
// 	go func() {
// 		replica, err := NewReplica(port)
// 		if err != nil {
// 			t.Fatal("Server failed to start:", err)
// 		}
// 		fmt.Println("Server started", replica.replica_id)
// 	}()

// 	// Give the server some time to start
// 	time.Sleep(time.Second)

// 	// Connect the client
// 	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("localhost: %d", port))
// 	if err != nil {
// 		t.Fatal("Failed to dial server:", err)
// 	}
// 	// defer client.Close()

// 	// Perform a test call
// 	args := Message{
// 		Type:        MsgPropose,
// 		OperationID: 1,
// 		Op: &Operation{
// 			op_type:   "test",
// 			key:       "test",
// 			value:     "test",
// 			timestamp: time.Now(),
// 		},
// 	} // fill this with actual arguments
// 	reply := Message{}
// 	err = client.Call("Replica.HAndleOperation", args, &reply)
// 	if err != nil {
// 		t.Fatal("RPC call failed:", err)
// 	}
// }
