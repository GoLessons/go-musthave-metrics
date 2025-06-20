package main

import (
	"github.com/GoLessons/go-musthave-metrics/internal/server"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	return http.ListenAndServe(`:8080`, server.InitRouter())
}
