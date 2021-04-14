package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const ( //Enter your PostgreSQL info,or change sql.Open parameters when using a different SQL management system
	DB_HOST     = ""
	DB_PORT     = ""
	DB_USERNAME = ""
	DB_PASSWORD = ""
	DB_NAME     = ""
)

var db *sql.DB

type Note struct {
	Id   int    `json:"id"`
	Note string `json:"note"`
}
type SuccessFlag struct {
	Success bool `json:"done"`
}

//Starts up the server the sets up handlers for requests
func main() {
	PORT := os.Getenv("PORT")
	var err error
	router := mux.NewRouter()
	db, err = sql.Open("postgres", fmt.Sprintf("dbname=%s user=%s password=%s host=%s port=%s",
		DB_NAME, DB_USERNAME, DB_PASSWORD, DB_HOST, DB_PORT))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	router.HandleFunc("/notes", func(w http.ResponseWriter, r *http.Request) {
		w.Write(getNotes())
	})
	router.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		var data []byte
		var jsonInterface struct {
			Note string `json:"Note"`
		}
		length, err := strconv.Atoi(r.Header.Get("Content-Length"))
		if err != nil {
			log.Fatal(err)
		}
		data = make([]byte, length)
		r.Body.Read(data)
		r.Body.Close()
		err = json.Unmarshal(data, &jsonInterface)
		if err != nil {
			log.Fatal(err)
		}
		addNote(jsonInterface.Note)
		var success SuccessFlag
		success.Success = true
		jsonData, err := json.Marshal(success)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(jsonData)
	})
	router.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		var data []byte
		var jsonInterface struct {
			Id int `json:"id"`
		}
		length, err := strconv.Atoi(r.Header.Get("Content-Length"))
		if err != nil {
			log.Fatal(err)
		}
		data = make([]byte, length)
		r.Body.Read(data)
		r.Body.Close()
		err = json.Unmarshal(data, &jsonInterface)
		if err != nil {
			log.Fatal(err)
		}
		deleteNote(jsonInterface.Id)
		var success SuccessFlag
		success.Success = true
		jsonData, err := json.Marshal(success)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(jsonData)
	})
	err = http.ListenAndServe(":"+PORT, router)
	if err != nil {
		log.Fatal(err)
	}
}

//Recieves notes from DB
func getNotes() []byte {
	var notes []Note
	var count int
	err := db.QueryRow("select count(*) from notes").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("select * from notes")
	if err != nil {
		log.Fatal(err)
	}
	notes = make([]Note, count)
	for i := 0; rows.Next(); i++ {
		var note Note
		err = rows.Scan(&note.Id, &note.Note)
		if err != nil {
			log.Fatal(err)
		}
		note.Note = strings.TrimSpace(note.Note)
		notes[i].Id = note.Id
		notes[i].Note = note.Note
	}
	if err = rows.Close(); err != nil {
		log.Fatal(err)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	jsonData, err := json.Marshal(notes)
	if err != nil {
		log.Fatal(err)
	}
	return jsonData
}

//Adds a note to DB
func addNote(note string) {
	_, err := db.Exec("INSERT INTO notes (note) VALUES($1)", note)
	if err != nil {
		log.Fatal(err)
	}
}

//Deletes a note from DB using its ID
func deleteNote(id int) {
	_, err := db.Exec("delete from notes where id=$1", id)
	if err != nil {
		log.Fatal(err)
	}
}
