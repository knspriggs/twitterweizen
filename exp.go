package main

import (
  "fmt"
  "strconv"
)

func increaseVote(value []byte) ([]byte) {
  val := string(value)
  int_value, _ := strconv.ParseInt(val, 10, 0)
  int_value++
  return []byte(strconv.FormatInt(int_value, 10))
}

func main() {
  value := []byte("0")
  fmt.Println(string(increaseVote(value)))
}
