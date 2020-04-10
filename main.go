package main

import (
	"log"

	"github.com/omerkaya1/grpc-file-stream/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Println(err)
	}
}
