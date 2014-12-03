package main

import (
	"github.com/boltdb/bolt"
	ustream "github.com/knspriggs/twitter-user-stream"
	"log"
	"strconv"
	"strings"
)

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
