package main

import (
	"log"

	"github.com/m110/cort/rpc"
)

func main() {
	client := rpc.NewClient("ExampleService")
	err := client.Connect()
	if err != nil {
		log.Fatal(err)
	}

	r, err := client.Call("example.method")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Got response:", r)
}
