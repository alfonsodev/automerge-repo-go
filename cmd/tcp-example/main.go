package main

import (
	"flag"
	"fmt"
	"net"

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

	r := repo.New()
	if *listenAddr != "" {
		ln, err := net.Listen("tcp", *listenAddr)
		if err != nil {
			panic(err)
		}
		fmt.Printf("listening on %s with repo %s\n", *listenAddr, r.ID)
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("accept error:", err)
				continue
			}
			go func(c net.Conn) {
				defer c.Close()
				remote, err := repo.Handshake(c, r.ID, repo.Incoming)
				if err != nil {
					fmt.Println("handshake error:", err)
					return
				}
				fmt.Println("connected peer", remote)
			}(conn)
		}
	} else if *connectAddr != "" {
		conn, err := net.Dial("tcp", *connectAddr)
		if err != nil {
			panic(err)
		}
		fmt.Printf("connecting to %s from repo %s\n", *connectAddr, r.ID)
		remote, err := repo.Handshake(conn, r.ID, repo.Outgoing)
		if err != nil {
			panic(err)
		}
		fmt.Println("connected to peer", remote)
	}
}
