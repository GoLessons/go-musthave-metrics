package main

import (
	"log"
	"os"
)

func main() {
	os.Exit(0)
	log.Fatal("fatal")
	log.New(os.Stderr, "", 0).Fatal("fatal-method")
}
