package main

import (
	"github.com/boltdb/bolt"
	ustream "github.com/knspriggs/twitter-user-stream"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var db *bolt.DB

type QuestionList struct {
	Questions []Question
}

type Question struct {
	User    string
	Id_str  string
	Text    string
	Yes     int64
	No      int64
	Yes_str string
	No_str  string
}

// Pull requests off the usteam channel
func GetRequests(requestsChannel chan *ustream.Tweet) {
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
			log.Printf("Request received, sending to be parsed")
			go ParseRequest(t, requestsChannel)
		}
	}
}

// Verifies if request is valid, or just another tweet on the user's stream
func ParseRequest(req *ustream.Tweet, requestsChannel chan *ustream.Tweet) {
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

	if question_tweet {
		requestsChannel <- req
	} else if exists(req.In_reply_to_status_id_str) {
		requestsChannel <- req
	} else {
		log.Printf("These are not the tweets you are looking for")
	}
}

// Deals with requests that are deemed valid
func HandleValidRequests(requestsChannel chan *ustream.Tweet) {
	log.Printf("Starting handler loop...")
	var req *ustream.Tweet
	for {
		req = <-requestsChannel
		log.Printf("Parsed request taken from channel: %#v", req)
		if req.In_reply_to_status_id_str == "null" || req.In_reply_to_status_id_str == "" {
			NewQuestion(req)
		} else {
			NewVote(req)
		}
	}
}

// Index handler for http requests to '/'
func indexHandler(w http.ResponseWriter, r *http.Request) {
	params := getIndexData()
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	t.Execute(w, params)
}

// Help handler for http requests to '/help'
func helpHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/help.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	t.Execute(w, nil)
}

func main() {
	printBeer()

	var err error
	db, err = bolt.Open("twitterweizen.db", 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	requestsChannel := make(chan *ustream.Tweet, 50)
	go GetRequests(requestsChannel)
	go HandleValidRequests(requestsChannel)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/help", helpHandler)
	http.ListenAndServe(":8080", nil)
}
