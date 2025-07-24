package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/example/automerge-repo-go/repo"
	automerge "github.com/automerge/automerge-go"
)

var (
	docHandle *repo.DocumentHandle
	repoHandle *repo.RepoHandle
	peers  []repo.RepoID
	mu     sync.Mutex
)

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "set", Description: "Set a key/value pair"},
		{Text: "get", Description: "Get a value by key"},
		{Text: "view", Description: "View the document as JSON"},
		{Text: "exit", Description: "Exit the application"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func executor(in string) {
	parts := strings.Fields(in)
	if len(parts) == 0 {
		return
	}
	switch parts[0] {
	case "set":
		if len(parts) != 3 {
			fmt.Println("usage: set <key> <value>")
			return
		}
		err := docHandle.WithDocMut(func(doc *automerge.Doc) error {
			return doc.RootMap().Set(parts[1], parts[2])
		})
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		mu.Lock()
		localPeers := append([]repo.RepoID(nil), peers...)
		mu.Unlock()
		for _, p := range localPeers {
			_ = repoHandle.SyncDocument(p, docHandle.DocID())
		}
	case "get":
		if len(parts) != 2 {
			fmt.Println("usage: get <key>")
			return
		}
		var val interface{}
		docHandle.WithDoc(func(doc *automerge.Doc) {
			v, err := doc.RootMap().Get(parts[1])
			if err != nil {
				fmt.Println("error:", err)
				return
			}
			val, err = automerge.As[interface{}](v)
			if err != nil {
				fmt.Println("error:", err)
				return
			}
		})
		fmt.Println(val)
	case "view":
		var data map[string]interface{}
		docHandle.WithDoc(func(doc *automerge.Doc) {
			var err error
			data, err = automerge.As[map[string]interface{}](doc.Root())
			if err != nil {
				fmt.Println("error:", err)
				return
			}
		})
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Println(string(b))
	case "exit":
		os.Exit(0)
	default:
		fmt.Println("unknown command")
	}
}

func main() {
	listenAddr := flag.String("listen", "", "address to listen on")
	connectAddr := flag.String("connect", "", "address to connect to")
	flag.Parse()

	if *listenAddr == "" && *connectAddr == "" {
		fmt.Println("specify -listen or -connect")
		return
	}

	r := repo.New()
	repoHandle = repo.NewRepoHandle(r)
	docHandle = r.NewDocHandle()
	docHandle.WithDocMut(func(doc *automerge.Doc) error {
		return doc.RootMap().Set("greeting", "hello")
	})

	go func() {
		for msg := range repoHandle.Inbox {
			fmt.Printf("\n< received %s message for doc %s\n> ", msg.Type, msg.DocumentID)
		}
	}()

	if *listenAddr != "" {
		go listen(*listenAddr)
	}

	if *connectAddr != "" {
		go connect(*connectAddr)
	}

	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("> "),
		prompt.OptionTitle("tcp-example"),
	)
	p.Run()
}

func listen(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("listen error:", err)
		os.Exit(1)
	}
	fmt.Printf("listening on %s with repo %s\n", ln.Addr().String(), repoHandle.Repo.ID)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		go handleConnection(conn, repo.Incoming)
	}
}

func connect(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("dial error:", err)
		return
	}
	handleConnection(conn, repo.Outgoing)
}

func handleConnection(conn net.Conn, dir repo.ConnDirection) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	lp, remote, err := repo.Connect(ctx, conn, repoHandle.Repo.ID, dir)
	if err != nil {
		fmt.Println("handshake error:", err)
		conn.Close()
		return
	}
	_ = repoHandle.AddConn(remote, lp)
	mu.Lock()
	peers = append(peers, remote)
	mu.Unlock()
	repoHandle.SyncAll(remote)
}