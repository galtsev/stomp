package main

import (
	"github.com/galtsev/stomp/server"
)

func main() {
	srv := server.NewServer()
	srv.ListenAndServe("localhost:1620")
}
