package dbs

import (
	_ "database/sql"
//	"fmt"
	"log"
	"encoding/json"
//	"io/ioutil"
//	"os"
	
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var schema = `
CREATE TABLE IF NOT EXISTS contactlist (
    id Serial PRIMARY KEY,
    name text
);
CREATE TABLE IF NOT EXISTS contact (
    id Serial PRIMARY KEY,
    first_name text,
    last_name text,
    email text,
    contactlist_id int REFERENCES contactlist(id) ON DELETE RESTRICT
);`

type Contact struct {
	Id int             `json:"-"`
	First_name string  `json:"first_name"`
	Last_name string   `json:"last_name"`
	Email string       `json:"email"`
	Contactlist_id int `json:"-"`
}

type ContactList struct {
	Id int             `json:"-"`
	Name string        `json:"name"`
	Contacts []Contact `json:"contacts"`
}

func (cl *ContactList) LoadFromJSON(json_str []byte) error {
	b := []byte(json_str)
	err := json.Unmarshal(b, &cl)
	if err != nil {
		log.Fatal(err)
	}

	return err
}

func (cl *ContactList) Save(db *sqlx.DB) error {
	namedStmt, err := db.PrepareNamed("INSERT INTO contactlist (name) VALUES (:name) RETURNING id")
	if err != nil {
		log.Fatal(err)
		return err
	}
	var contactlist_id int
	err = namedStmt.Get(&contactlist_id, cl)
	if err != nil {
		log.Fatal(err)
		return err
	}
	
	tx := db.MustBegin()
	stmt, _ := tx.Preparex(pq.CopyIn("contact", "first_name", "last_name", "email", "contactlist_id"))
	for _, contact := range cl.Contacts {
		stmt.Exec(contact.First_name, contact.Last_name, contact.Email, contactlist_id)
	}
	_, err = stmt.Exec()
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

func (cl *ContactList) LoadFromDB(db *sqlx.DB, id int) error {
	err := db.Get(cl, `SELECT * FROM contactlist WHERE id=$1`, id)
	if err != nil {
		log.Fatal(err)

		return err
	}
	return cl.loadContactsFromDB(db)
}

func (cl *ContactList) loadContactsFromDB(db *sqlx.DB) error {
	if cl.Id > 0 {
		err := db.Select(&cl.Contacts, `SELECT * FROM contact WHERE contactlist_id=$1`, cl.Id)
		if err != nil {
			log.Fatal(err)
		}

		return err
	}

	return nil
}


func Connect() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", "user=pqgotest password=postgres dbname=pqgotest sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	return db, err
}


func CreateTables(db *sqlx.DB) {
	db.MustExec(schema)
}

func GetContactListsFromDB(db *sqlx.DB) ([]ContactList, error) {
	contactlists := []ContactList{}
	err := db.Select(&contactlists, `SELECT * FROM contactlist ORDER BY name,id`)
	if err != nil {
		log.Fatal(err)
	}
	
	return contactlists, err
}

// func main() {
// 	db, err := connect()
// 	if err != nil {
// 		os.Exit(1)
// 	}

// 	createTables(db)

// 	tx := db.MustBegin()
// 	tx.MustExec("DELETE FROM contact")
// 	tx.MustExec("DELETE FROM contactlist")
// 	tx.Commit()

// 	contactListJson, e := ioutil.ReadFile("./contactlist.json")
// 	if e != nil {
// 		fmt.Printf("File error: %v\n", e)
// 		os.Exit(1)
// 	}
	
// 	contactList := new(ContactList)
// 	contactList.loadFromJSON(contactListJson)
// 	contactList.save(db)

// 	lists, _ := getContactListsFromDB(db)

// 	fmt.Println(lists[0])
// 	lists[0].loadContactsFromDB(db)
// 	fmt.Println(lists[0])
	
// 	contactList2 := new(ContactList)
// 	contactList2.loadFromDB(db, lists[0].Id)
// 	fmt.Println(*contactList2)
// }
