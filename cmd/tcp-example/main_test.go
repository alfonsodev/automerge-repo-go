package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/example/automerge-repo-go/repo"
	automerge "github.com/automerge/automerge-go"
)

func TestExecutor(t *testing.T) {
	r := repo.New()
	repoHandle = repo.NewRepoHandle(r)
	docHandle = r.NewDocHandle()
	docHandle.WithDocMut(func(doc *automerge.Doc) error {
		return doc.RootMap().Set("greeting", "hello")
	})

	oldStdout := os.Stdout
	ro, w, _ := os.Pipe()
	os.Stdout = w

	executor("get greeting")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, ro)

	expected := "hello"
	if !strings.Contains(buf.String(), expected) {
		t.Errorf("expected output to contain %q, but got %q", expected, buf.String())
	}
}
