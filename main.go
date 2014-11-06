package main

// Initial imports, not sure if there is a conflict of types between the two twitter-go libraries
import (
	"github.com/boltdb/bolt"
	ustream "github.com/knspriggs/twitter-user-stream"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var db *bolt.DB

// Server loop to wait for requests
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

// ParseRequest: Adds to channel if request is valid, exits otherwise
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
		//add to channel to be stored
		requestsChannel <- req
	} else if string(req.In_reply_to_status_id_str) != "null" || string(req.In_reply_to_status_id_str) != "" {
		requestsChannel <- req
	} else {
		log.Printf("These are not the tweets you are looking for")
	}
}

// HandQuestions: Loop that takes requests from request channel to process
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

func NewQuestion(req *ustream.Tweet) {
	log.Printf("Logging tweet to db: %s : %s", req.User.Screen_name, req.Id_str)
	err := db.Update(func(tx *bolt.Tx) error {
		m, err := tx.CreateBucketIfNotExists([]byte("keys"))
		if err != nil {
			log.Printf("create bucket: %s", err)
		}
		m.Put([]byte(req.Id_str), []byte(""))

		b, err := tx.CreateBucketIfNotExists([]byte(req.Id_str))
		if err != nil {
			log.Printf("create bucket: %s", err)
		}
		b.Put([]byte("text"), []byte(req.Text))
		b.Put([]byte("yes"), []byte("0"))
		b.Put([]byte("no"), []byte("0"))

		return nil
	})
	if err != nil {
		log.Printf("Errpr adding question: %s", err)
	}
}

func NewVote(req *ustream.Tweet) {
	log.Printf("Registering a new vote")
	var vote byte
	if contains(strings.Split(req.Text, " "), "yes") {
		vote = 't'
	} else if contains(strings.Split(req.Text, " "), "no") {
		vote = 'f'
	} else {
		vote = 'q'
	}

	if vote != 'q' {
		err := db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(req.In_reply_to_status_id_str))
			if vote == 't' {
				b.Put([]byte("yes"), increaseVote(b.Get([]byte("yes"))))
			} else {
				b.Put([]byte("no"), increaseVote(b.Get([]byte("no"))))
			}
			return nil
		})
		if err != nil {
			log.Printf("Error adding vote: %s", err)
		}
	}
}

func increaseVote(value []byte) []byte {
	val := string(value)
	int_value, _ := strconv.ParseInt(val, 10, 0)
	int_value++
	return []byte(strconv.FormatInt(int_value, 10))
}

func PrintStats() {
	for {
		time.Sleep(10 * time.Second)
		log.Printf("Printing stats:")
		db.View(func(tx *bolt.Tx) error {
			m := tx.Bucket([]byte("keys"))
			if m == nil {
				log.Printf("Bucket %q not found!", []byte("keys"))
			} else {
				m.ForEach(func(k, v []byte) error {
					b := tx.Bucket([]byte(k))
					b.ForEach(func(k, v []byte) error {
						log.Printf("key=%s, value=%s\n", k, v)
						return nil
					})
					return nil
				})
			}
			return nil
		})
	}
}

func main() {
	printBeer()
	log.Printf("Start me up!")

	//TEMP
	os.Setenv("TWITTER_USER_NAME", "kristianspriggs")

	var err error

	db, err = bolt.Open("twitterweizen.db", 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	requestsChannel := make(chan *ustream.Tweet, 50)

	go GetRequests(requestsChannel)
	//go PrintStats()
	HandleQuestions(requestsChannel)

	defer db.Close()
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
