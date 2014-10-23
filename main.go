package main

// Initial imports, not sure if there is a conflict of types between the two twitter-go libraries
import (
	redis "github.com/hoisie/redis"
	ustream "github.com/knspriggs/twitter-user-stream"
	"log"
	"os"
	"strings"
)

var client redis.Client

type Question struct {
	Tweet_id string
	User     *ustream.User
	Text     string
	Votes    []*Vote
}

type Vote struct {
	Text  string
	Count int64
}

// Server loop to wait for requests
func GetRequests(cQuestion chan *Question) {
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
			go ParseRequest(t, cQuestion)
		}
	}
}

// ParseRequest: Adds to channel if request is valid, exits otherwise
func ParseRequest(req *ustream.Tweet, cQuestion chan *Question) {
	// Get environment variables for config
	handle := os.Getenv("TWITTER_USER_NAME")

	question_flags := []string{"#yesno", "#yesorno"}
	var question_pieces []string
	var question string

	question_tweet := false

	//---- ERROR SOMEWHERE IN HERE
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
	//----

	log.Printf("%s : %s : %d - %b", req.User.Screen_name, req.In_reply_to_status_id_str, req.Id, question_tweet)

	if question_tweet {
		//add to channel to be posted!
		p := new(Question)
		(*p).Text = question
		(*p).User = req.User
		(*p).Tweet_id = req.In_reply_to_status_id_str
		cQuestion <- p
	} else if string(req.In_reply_to_status_id_str) != "null" {
		if ok, _ := client.Exists(req.In_reply_to_status_id_str); ok {
			log.Printf("We found a tweet!")
		}
	} else {
		log.Println("These are not the tweets you are looking for")
	}
}

// HandRequests: Loop that takes requests from request channel to process
func HandleRequests(cQuestion chan *Question) {
	log.Println("Starting handler loop...")
	var q *Question
	for {
		q = <-cQuestion
		log.Println("Parsed request taken from channel")
		go PostToServer(q)
	}
}

func PostToServer(question *Question) {
	log.Printf("DO SOMETHING CRAZY!")
}

func main() {
	printBeer()
	log.Printf("Start me up!")

	//TEMP
	os.Setenv("TWITTER_USER_NAME", "@kristianspriggs")
	os.Setenv("REDIS_HOST_ADDRESS", "172.17.0.3:6379")

	// Redis DB connections
	client.Addr = os.Getenv("REDIS_HOST_ADDRESS")

	// TODO: Play around with buffer size
	cQuestion := make(chan *Question, 50)

	// Start the server loop
	go GetRequests(cQuestion)
	// Handle requests
	HandleRequests(cQuestion)

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
