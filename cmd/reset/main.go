package main

import (
	"flag"
	"fmt"
	"os"

	resettool "github.com/GoLessons/go-musthave-metrics/cmd/reset/internal"
)

func main() {
	root := "."
	flags := flag.NewFlagSet("reset", flag.ExitOnError)
	flags.StringVar(&root, "root", ".", "root directory to scan")
	_ = flags.Parse(os.Args[1:])

	if err := resettool.Run(root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
