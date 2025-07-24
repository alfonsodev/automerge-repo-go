
package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestExampleMain(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = []string{"example", "new"}
	main()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if len(buf.String()) == 0 {
		t.Errorf("expected output, but got empty string")
	}
}
