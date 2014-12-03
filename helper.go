package main

import (
	"log"
)

//array contains
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Because why not?
func printBeer() {
	log.Printf("\n\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n\n", "     oOOOOOo",
		"    ,|    oO", "   //|     |", "   \\ |     |",
		"    `|     |", "     `-----`", "  Twitterweizen",
		"  by: knspriggs")
}
