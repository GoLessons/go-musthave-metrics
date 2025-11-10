package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestExitCheck(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, ExitCheckAnalyzer,
		"allowed",
		"main_outside",
		"forbidden",
		"panicmain",
		"paniccases",
		"dotallowed",
		"dotforbidden",
	)
}
