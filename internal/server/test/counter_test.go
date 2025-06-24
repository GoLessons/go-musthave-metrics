package test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSomeCounter(t *testing.T) {
	var I *tester
	I = NewTester()
	defer I.Shutdown()
	I.Test(t)

	resp, err := I.Get("/update/counter/qwerty/100")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
