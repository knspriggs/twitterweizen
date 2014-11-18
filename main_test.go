package main

import (
	"testing"
)

func TestGenerateString(t *testing.T) {
	s := generateString(5, true)
	if s != "+++++" {
		t.Error("Expect '+++++', got %s", s)
	}
}
