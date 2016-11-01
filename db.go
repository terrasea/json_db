package main

import (
	_ "database/sql"
	"fmt"
	"log"
	"encoding/json"
	"io/ioutil"
	"os"
	
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

func (c *Contacts) createFromJSON(json_str []byte) error {
	b := []byte(json_str)
	err := json.Unmarshal(b, &c)
	if err != nil {
		log.Fatal(err)
	}

	return err
}

func (c *Contacts) save(db *sqlx.DB) error {
	tx := db.MustBegin()
	stmt, _ := tx.Preparex(pq.CopyIn("contact", "first_name", "last_name", "email"))
	
	for _, contact := range c.Contacts {
		tx.Exec(contact.First_name, contact.Last_name, contact.Email)
	}
	_, err := stmt.Exec()
	if err != nil {
		log.Fatal(err)
		return err
	}
	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
		return err
	}
	
	tx.Commit()
	
	return nil
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


func main() {
	db, err := connect()
	if err != nil {
		os.Exit(1)
	}

	createTables(db)
	contactsJson, e := ioutil.ReadFile("./contacts.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	tx := db.MustBegin()
	tx.MustExec("DELETE FROM contact")
	tx.Commit()
	
	contacts := new(Contacts)
	
	contacts.createFromJSON(contactsJson)
	
	contacts.save(db)
	

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
