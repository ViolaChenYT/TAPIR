package main

import (
	. "github.com/ViolaChenYT/TAPIR/tapir_kv"
)

func main() {
	app := NewTapirApp(nil)
	app.Start()
	row := make(map[string][]byte)
	row["name"] = []byte("ruyu")
	row["netid"] = []byte("ry9811")
	app.Insert("123", "456", row)
	app.Read("123", "456", []string{"name"})
	app.Commit()
}
