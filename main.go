package main

import (
	"log"

	"github.com/ViolaChenYT/TAPIR/common"
	. "github.com/ViolaChenYT/TAPIR/tapir_kv"
)

func main() {
	app := NewTapirApp(common.GetConfigA())
	app.Start()
	row := make(map[string][]byte)
	row["name"] = []byte("ruyu")
	row["netid"] = []byte("ry9811")
	app.Insert("123", "456", row)
	app.Read("123", "456", []string{"name"})
	app.Commit()

	app.Start()
	val, err := app.Read("123", "456", []string{"name"})
	log.Println(val, err)
}
