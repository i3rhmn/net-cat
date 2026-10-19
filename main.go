package main

import (
	"fmt"
	"os"

	"net-cat/tcpchat"
)

func main() {
	port := "8989"

	if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	if len(os.Args) == 2 {
		port = os.Args[1]
	}

	if err := tcpchat.RunServer(port); err != nil {
		fmt.Println("Server error:", err)
	}
}
