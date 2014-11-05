package main

import (
  ustream "github.com/knspriggs/twitter-user-stream"
  "github.com/boltdb/bolt"
  "log"
  "os"
  "strings"
  "time"
)

var db bolt.DB


func GetRequests() {
  for {
    ustreamClient := ustream.NewUStreamClient()
    httpResp, err := ustreamClient.Connect()
    if err != nil {
      log.Println(err)
      continue
    }
    cTweets := ustreamClient.ReadStream(httpResp)
    var t *ustream.Tweet
    for {
      t = <-cTweets
      log.Println("Request received, sending to be parsed")
      go ParseRequest(t)
    }
  }
}

func ParseRequest(req *ustream.Tweet,) {
  // Get environment variables for config
  handle := os.Getenv("TWITTER_USER_NAME")

  question_flags := []string{"#yesno", "#yesorno"}
  var question_pieces []string
  question_tweet := false

  arr := strings.Split(req.Text, " ")
  if req.User.Screen_name == handle {
    for k := 0; k < len(arr); k++ {
      if arr[k] == handle {
        //move on
      } else if arr[k][0] == '#' {
        if contains(question_flags, arr[k]) {
          question_tweet = true
        }
      } else {
        question_pieces = append(question_pieces, arr[k])
      }
    }
  }

  log.Printf("%s : %s : %d - %b", req.User.Screen_name, req.In_reply_to_status_id_str, req.Id, question_tweet)

  if question_tweet {
    StoreQuestion(req)
  } else if string(req.In_reply_to_status_id_str) != "null" {
    StoreVote(req)
  } else {
    log.Println("These are not the tweets you are looking for")
  }
}

func StoreQuestion(req *ustream.Tweet) {

}

func StoreVote(req *ustream.Tweet) {

}


func main() {
  printBeer()

  os.Setenv("TWITTER_USER_NAME", "@kristianspriggs")

  db, err := bolt.Open("twitterweizen.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  //requestsChannel := make(chan *ustream.Tweet, 50)

  go GetRequests()
}

// Helper methods
func contains(s []string, e string) bool {
  for _, a := range s {
    if a == e {
      return true
    }
  }
  return false
}

func printBeer() {
  log.Printf("\n\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n\n", "     oOOOOOo",
    "    ,|    oO", "   //|     |", "   \\ |     |",
    "    `|     |", "     `-----`", "  Twitterweizen",
    "  by: knspriggs")
}
