package main

import (
	"fmt"

	"github.com/example/automerge-repo-go/repo"
)

func main() {
	store := &repo.FsStore{Dir: "data"}
	r := repo.NewWithStore(store)

	doc := r.NewDoc()
	if err := doc.Set("greeting", "hello"); err != nil {
		panic(err)
	}
	if err := r.SaveDoc(doc.ID); err != nil {
		panic(err)
	}

	loaded, err := r.LoadDoc(doc.ID)
	if err != nil {
		panic(err)
	}
	value, _ := loaded.Get("greeting")
	fmt.Println("loaded:", value)
}
