package main

import (
	"fmt"
	"os"

	"github.com/example/automerge-repo-go/repo"
	"github.com/google/uuid"
)

func usage() {
	fmt.Println(`Usage:
  example new                       create a new document
  example list                      list document IDs
  example set <id> <key> <value>    set a value in a document
  example get <id> <key>            get a value from a document`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	store := &repo.FsStore{Dir: "data"}
	r := repo.NewWithStore(store)

	switch os.Args[1] {
	case "new":
		doc := r.NewDoc()
		if err := r.SaveDoc(doc.ID); err != nil {
			panic(err)
		}
		fmt.Println(doc.ID)
	case "list":
		ids, err := store.List()
		if err != nil {
			panic(err)
		}
		for _, id := range ids {
			fmt.Println(id)
		}
	case "set":
		if len(os.Args) != 5 {
			usage()
			return
		}
		id, err := uuid.Parse(os.Args[2])
		if err != nil {
			panic(err)
		}
		key, value := os.Args[3], os.Args[4]
		doc, err := r.LoadDoc(id)
		if err != nil {
			panic(err)
		}
		if err := doc.Set(key, value); err != nil {
			panic(err)
		}
		if err := r.SaveDoc(id); err != nil {
			panic(err)
		}
	case "get":
		if len(os.Args) != 4 {
			usage()
			return
		}
		id, err := uuid.Parse(os.Args[2])
		if err != nil {
			panic(err)
		}
		key := os.Args[3]
		doc, err := r.LoadDoc(id)
		if err != nil {
			panic(err)
		}
		if v, ok := doc.Get(key); ok {
			fmt.Println(v)
		}
	default:
		usage()
	}
}
