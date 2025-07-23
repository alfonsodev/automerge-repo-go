package main

import (
	"flag"
	"fmt"
	"net"
	"os"

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
		fmt.Printf("listening on %s with repo %s\n", *listenAddr, handle.Repo.ID)
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("accept error:", err)
				continue
			}
			go func(c net.Conn) {
				lp, remote, err := repo.Connect(c, handle.Repo.ID, repo.Incoming)
				if err != nil {
					fmt.Println("handshake error:", err)
					c.Close()
					return
				}
				handle.AddConn(remote, lp)
				handle.SyncAll(remote)
			}(conn)
		}
	} else if *connectAddr != "" {
		conn, err := net.Dial("tcp", *connectAddr)
		if err != nil {
			fmt.Println("dial error:", err)
			return
		}
		lp, remote, err := repo.Connect(conn, handle.Repo.ID, repo.Outgoing)
		if err != nil {
			fmt.Println("handshake error:", err)
			return
		}
		handle.AddConn(remote, lp)
		handle.SyncAll(remote)
	}
}
