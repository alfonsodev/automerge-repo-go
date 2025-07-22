package main

import (
	"fmt"
	"github.com/example/automerge-repo-go/repo"
)

func main() {
	r := repo.New()
	doc := r.NewDoc()
	fmt.Println("new repo:", r.ID)
	fmt.Println("new document:", doc.ID)
}
