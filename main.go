package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

var (
	cache    map[string]string
	mu       *sync.Mutex
	sockPath = "/tmp/cache.sock"
	ttl      = time.Minute
)

func init() {
	cache = make(map[string]string)
	mu = &sync.Mutex{}
}

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		close(quit)

		log.Println("Removing socket: ", os.RemoveAll(sockPath))
		os.Exit(0)
	}()

	go func() {
		for {
			time.Sleep(ttl)
			mu.Lock()
			cache = make(map[string]string)
			mu.Unlock()
		}
	}()

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept fail: %v", err)
			continue
		}

		bs, err := io.ReadAll(conn)
		if err != nil {
			log.Println("ReadAll fail: %v", err)
			continue
		}

		split := strings.Split(string(bs), " ")

		switch len(split) {
		case 2:
			mu.Lock()
			cache[split[0]] = split[1]
			mu.Unlock()
		case 1:
			fmt.Fprintf(conn, cache[split[0]])
		}
	}
}
