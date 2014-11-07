package main

// Initial imports, not sure if there is a conflict of types between the two twitter-go libraries
import (
	"github.com/boltdb/bolt"
	ustream "github.com/knspriggs/twitter-user-stream"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
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

// Register a new question
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
		b.Put([]byte("user"), []byte(req.User.Screen_name))
		b.Put([]byte("text"), []byte(req.Text))
		b.Put([]byte("yes"), []byte("0"))
		b.Put([]byte("no"), []byte("0"))

		return nil
	})
	if err != nil {
		log.Printf("Error adding question: %s", err)
	}
}

// Register a new vote
func NewVote(req *ustream.Tweet) {
	log.Printf("Registering a new vote")
	var vote byte
	if contains(strings.Split(req.Text, " "), "#yes") {
		vote = 't'
	} else if contains(strings.Split(req.Text, " "), "#no") {
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

// Update the vote data in the DB
func increaseVote(value []byte) []byte {
	val := string(value)
	int_value, _ := strconv.ParseInt(val, 10, 0)
	int_value++
	return []byte(strconv.FormatInt(int_value, 10))
}

// Get data neccesary to populate index page from DB
func getIndexData() *QuestionList {
	var list []Question
	db.View(func(tx *bolt.Tx) error {
		m := tx.Bucket([]byte("keys"))
		if m == nil {
			log.Printf("Bucket %q not found!", []byte("keys"))
		} else {
			m.ForEach(func(k, v []byte) error {
				b := tx.Bucket([]byte(k))
				user := b.Get([]byte("user"))
				text := b.Get([]byte("text"))
				yes, _ := strconv.ParseInt(string(b.Get([]byte("yes"))), 10, 0)
				no, _ := strconv.ParseInt(string(b.Get([]byte("no"))), 10, 0)
				list = append(list, Question{User: string(user), Id_str: string(k), Text: string(text), Yes: yes, No: no, Yes_str: generateString(yes, true), No_str: generateString(no, false)})
				return nil
			})
		}
		return nil
	})
	return &QuestionList{Questions: list}
}

// Index handler for http requests to '/'
func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("indexHandler invoked")
	params := getIndexData()
	t, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	t.Execute(w, params)
}

// Help handler for http requests to '/help'
func helpHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("helpHandler invoked")
	t, err := template.ParseFiles("help.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	t.Execute(w, nil)
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
	defer db.Close()

	requestsChannel := make(chan *ustream.Tweet, 50)
	go GetRequests(requestsChannel)
	//go PrintStats()
	go HandleValidRequests(requestsChannel)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/help", helpHandler)
	http.ListenAndServe(":8080", nil)
}

// ----- Helper methods
// Print DB stats for debugging
func PrintStats() {
	for {
		time.Sleep(10 * time.Second)
		log.Printf("Printing stats:")
		list := getIndexData()
		log.Printf("%#v", list)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// There has got to be an easier way to do this
func generateString(x int64, b bool) string {
	s := ""
	if b {
		for k := int64(0); k < x; k++ {
			s += "+"
		}
	} else {
		for k := int64(0); k < x; k++ {
			s += "-"
		}
	}
	return s
}

// Return if key exists in DB
func exists(key string) bool {
	var result bool
	db.View(func(tx *bolt.Tx) error {
		m := tx.Bucket([]byte("keys"))
		t := m.Get([]byte(key))
		if t != nil {
			result = true
		} else {
			result = false
		}
		return nil
	})
	return result
}

// Because why not?
func printBeer() {
	log.Printf("\n\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n\n", "     oOOOOOo",
		"    ,|    oO", "   //|     |", "   \\ |     |",
		"    `|     |", "     `-----`", "  Twitterweizen",
		"  by: knspriggs")
}
