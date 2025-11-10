package main

import (
	"log"
	"os"
)

func main() {}

func init() {
	os.Exit(1) // want "вызов os.Exit вне функции main пакета main"
}

func helper() {
	log.Fatal("fatal") // want "вызов log.Fatal вне функции main пакета main"
}
