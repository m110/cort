package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/m110/cort/service"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	srv := service.NewService("ExampleService", "tcp://127.0.0.1:9999")

	go func() {
		<-c
		srv.Stop()
		os.Exit(0)
	}()

	err := srv.Start()
	if err != nil {
		log.Fatal(err)
	}
}
