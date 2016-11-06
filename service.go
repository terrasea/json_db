package main

import (
	"fmt"
	"io/ioutil"
	"os"
	
	"github.com/terrasea/json_db/dbs"
)

func main() {
	db, _ := dbs.Connect()

	dbs.CreateTables(db)
	
	tx := db.MustBegin()
 	tx.MustExec("DELETE FROM contact")
	tx.MustExec("DELETE FROM contactlist")
	tx.Commit()
	
	contactListJson, e := ioutil.ReadFile("./contactlist.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	contactsList := new(dbs.ContactList)
	contactsList.LoadFromJSON(contactListJson)
	contactsList.Save(db)

	lists, _ := dbs.GetContactListsFromDB(db)
	
	contactsList2 := new(dbs.ContactList)
	contactsList2.LoadFromDB(db, lists[0].Id)
	fmt.Println(*contactsList2)
}
