package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/example/automerge-repo-go/repo"
)

func main() {
	listenAddr := flag.String("listen", "", "address to listen on")
	connectAddr := flag.String("connect", "", "address to connect to")
	flag.Parse()

	if *listenAddr == "" && *connectAddr == "" {
		fmt.Println("specify -listen or -connect")
		return
	}

	handle := repo.NewRepoHandle(repo.New())
	doc := handle.Repo.NewDoc()
	_ = doc.Set("greeting", "hello")

	var (
		mu    sync.Mutex
		peers []repo.RepoID
	)

	go func() {
		for msg := range handle.Inbox {
			fmt.Printf("received %s message for doc %s\n", msg.Type, msg.DocumentID)
		}
	}()

	if *listenAddr != "" {
		ln, err := net.Listen("tcp", *listenAddr)
		if err != nil {
			fmt.Println("listen error:", err)
			os.Exit(1)
		}
		fmt.Printf("listening on %s with repo %s\n", ln.Addr().String(), handle.Repo.ID)
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					fmt.Println("accept error:", err)
					continue
				}
				go func(c net.Conn) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					lp, remote, err := repo.Connect(ctx, c, handle.Repo.ID, repo.Incoming)
					if err != nil {
						fmt.Println("handshake error:", err)
						c.Close()
						return
					}
					_ = handle.AddConn(remote, lp)
					mu.Lock()
					peers = append(peers, remote)
					mu.Unlock()
					handle.SyncAll(remote)
				}(conn)
			}
		}()
	}

	if *connectAddr != "" {
		conn, err := net.Dial("tcp", *connectAddr)
		if err != nil {
			fmt.Println("dial error:", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		lp, remote, err := repo.Connect(ctx, conn, handle.Repo.ID, repo.Outgoing)
		if err != nil {
			fmt.Println("handshake error:", err)
			return
		}
		_ = handle.AddConn(remote, lp)
		mu.Lock()
		peers = append(peers, remote)
		mu.Unlock()
		handle.SyncAll(remote)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("commands: set <key> <value>, get <key>, exit")
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		parts := strings.Fields(scanner.Text())
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "set":
			if len(parts) != 3 {
				fmt.Println("usage: set <key> <value>")
				continue
			}
			if err := doc.Set(parts[1], parts[2]); err != nil {
				fmt.Println("error:", err)
				continue
			}
			mu.Lock()
			localPeers := append([]repo.RepoID(nil), peers...)
			mu.Unlock()
			for _, p := range localPeers {
				_ = handle.SyncDocument(p, doc.ID)
			}
		case "get":
			if len(parts) != 2 {
				fmt.Println("usage: get <key>")
				continue
			}
			if v, ok := doc.Get(parts[1]); ok {
				fmt.Println(v)
			}
		case "exit":
			return
		default:
			fmt.Println("unknown command")
		}
	}
}
