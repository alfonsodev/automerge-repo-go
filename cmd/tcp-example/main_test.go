package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTCPExampleMain(t *testing.T) {
	// Store original os.Args, os.Stdout, and os.Stdin
	oldArgs := os.Args
	oldStdout := os.Stdout
	oldStdin := os.Stdin
	defer func() {
		os.Args = oldArgs
		os.Stdout = oldStdout
		os.Stdin = oldStdin
	}()

	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run main in a goroutine as a listener
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd", "-listen", "127.0.0.1:0"}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		main()
	}()

	// Give the listener time to start
	time.Sleep(1 * time.Second)

	// Close the write end of the pipe and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the output from the listener
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Find the listening address from the output
	output := buf.String()
	lines := strings.Split(output, "\n")
	var listenAddr string
	for _, line := range lines {
		if strings.Contains(line, "listening on") {
			parts := strings.Split(line, " ")
			listenAddr = parts[2]
			break
		}
	}

	if listenAddr == "" {
		t.Fatal("could not find listening address in output")
	}

	// Now run another instance as a client
	clientInR, clientInW, _ := os.Pipe()
	os.Stdin = clientInR
	r, w, _ = os.Pipe()
	os.Stdout = w

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd", "-connect", listenAddr}
	wg.Add(1)
	go func() {
		defer wg.Done()
		main()
	}()

	// Give the client time to connect
	time.Sleep(1 * time.Second)

	// Write the "get" command to the client's stdin
	io.WriteString(clientInW, "get greeting\n")

	// Give the client time to process the command
	time.Sleep(1 * time.Second)

	// Close the write end of the pipe and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the output from the client
	var clientBuf bytes.Buffer
	io.Copy(&clientBuf, r)

	// Check for the expected output from the client
	expected := "hello"
	if !strings.Contains(clientBuf.String(), expected) {
		t.Errorf("expected client output to contain %q, but got %q", expected, clientBuf.String())
	}
}
