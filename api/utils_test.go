package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Util_RandomInt(t *testing.T) {
	num := RandomInt(1, 100)
	assert.Greater(t, num, 1)
	assert.Less(t, num, 100)
}

func Test_Util_RandomString(t *testing.T) {
	str := RandomString(5)
	assert.Len(t, str, 5)
}
