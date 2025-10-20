package main

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
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
