package main

import (
	"testing"
)

func TestContainsTrue(t *testing.T) {
	arr := []string{"yes", "no", "maybe"}
	if !contains(arr, "yes") {
		t.Error("Expected true, got false")
	}
}

func TestContainsFalse(t *testing.T) {
	arr := []string{"yes", "no", "maybe"}
	if contains(arr, "hello") {
		t.Error("Expected false, got true")
	}
}

func TestContainsEmpty(t *testing.T) {
	arr := []string{}
	if contains(arr, "no") {
		t.Error("Expected false, got true")
	}
}
