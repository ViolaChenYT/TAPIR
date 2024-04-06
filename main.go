package tapir

import (
	"IR"
	"fmt"
)

func Main() {
	fmt.Println("Hello, Tapir!")
	server := IR.NewServer(1)
	server.Start()
}
