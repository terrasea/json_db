package main

import (
	"github.com/lib/pq"
	_ "database/sql"
	"github.com/jmoiron/sqlx"
	"log"
	"fmt"
	"os"
	"encoding/json"
	"io/ioutil"
//	"github.com/gdey/sql2json"
)

var schema = `
CREATE TABLE IF NOT EXISTS contact (
    id Serial,
    first_name text,
    last_name text,
    email text
);`

type Contact struct {
	Id int            `json:"-"`
	First_name string `json:"first_name"`
	Last_name string  `json:"last_name"`
	Email string      `json:"email"`
}

type Contacts struct {
	Contacts []Contact `json:"contacts"`
}

func connect() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", "user=pqgotest dbname=pqgotest sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	return db, err
}


func createTables(db *sqlx.DB) {
	db.MustExec(schema)
}

func createContactsFromJSON(json_str string) ([]Contact, error) {
	b := []byte(json_str)
	var m Contacts
	err := json.Unmarshal(b, &m)
	if err != nil {
		log.Fatal(err)
	}

	return m.Contacts, err
}

func (c *Contact) save(tx *sqlx.Stmt) {
	tx.Exec(c.First_name, c.Last_name, c.Email)
}

func main() {
	db, err := connect()
	if err != nil {
		os.Exit(1)
	}

	createTables(db)
	file, e := ioutil.ReadFile("./contacts.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	tx := db.MustBegin()
	tx.MustExec("DELETE FROM contact")
	contacts, _ := createContactsFromJSON(string(file))
	stmt, _ := tx.Preparex(pq.CopyIn("contact", "first_name", "last_name", "email"))
	for _, contact := range contacts {
		contact.save(stmt)
	}
	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}
	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
	}
	
	tx.Commit()

	people := []Contact{}
	db.Select(&people, "SELECT * FROM contact ORDER BY email,id ASC")
	//jason, john := people[0], people[1]
	for _, contact := range people {
		contact_json, err := json.Marshal(contact)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", contact_json)
	}

	
}
