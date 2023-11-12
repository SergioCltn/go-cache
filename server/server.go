package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type CacheEntry struct {
	Key   string      `json:"key"`
	Value any         `json:"value"`
	TTL   time.Time   `json:"TTL"`
	mu    *sync.Mutex `json:"-"`
}
type CacheServer struct {
	cache       map[string]CacheEntry
	mu          sync.Mutex
	persistence Persistence
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewCacheServer(persistence Persistence) *CacheServer {
	context, cancel := context.WithCancel(context.Background())
	return &CacheServer{
		cache:       make(map[string]CacheEntry),
		persistence: persistence,
		ctx:         context,
		cancel:      cancel,
	}
}

func (cs *CacheServer) cleanupExpiredData() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cs.mu.Lock()
			for key, item := range cs.cache {
				if time.Now().After(item.TTL) {
					fmt.Printf("DELETING EXPIRED KEY: %s\n", key)
					delete(cs.cache, key)
				}
			}
			cs.mu.Unlock()

		case <-cs.ctx.Done():
			return
		}
	}
}

func (cs *CacheServer) handleConnection(conn net.Conn) {
	r := bufio.NewScanner(conn)

	for r.Scan() {
		command := r.Text()
		commandArray := strings.Fields(command)

		if len(commandArray) < 1 {
			conn.Write([]byte("Invalid command format\n"))
			continue
		}
		switch commandArray[0] {
		case "GET":
			if len(commandArray) != 2 {
				conn.Write([]byte("Invalid GET command format\n"))
				continue
			}
			fmt.Println("Get operation")
			cs.handleGet(conn, commandArray[1])
		case "SET":
			if len(commandArray) != 4 {
				conn.Write([]byte("Invalid SET command format\n"))
				continue
			}
			cs.handleSet(conn, commandArray[1], commandArray[2], commandArray[3])
		case "GETALL":
			cs.handleGetAll(conn)
		case "DELETE":
			cs.handleDelete(conn, commandArray[1])
		default:
			conn.Write([]byte("Invalid command lil nigga\n"))
		}
	}

}

func main() {
	li, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal("Error with connection")
		return
	}
	defer li.Close()
	persistence := newFilePersistence("storage.json")
	cs := NewCacheServer(persistence)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go cs.cleanupExpiredData()

	go func() {
		<-sigCh
		fmt.Println("\nShutting down server...")
		cs.cancel()
		li.Close()
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cs.mu.Lock()
				cs.persistence.Save(cs.cache)
				cs.mu.Unlock()

			case <-cs.ctx.Done():
				return
			}
		}
	}()

	fmt.Println("listening to port :8081")
	for {
		conn, err := li.Accept()
		if err != nil {
			// Check if the error is due to listener being closed (for graceful shutdown)
			select {
			case <-cs.ctx.Done():
				fmt.Println("Server has been gracefully shut down.")
				return
			default:
				log.Println("Error accepting connection:", err)
			}
			continue
		}

		go cs.handleConnection(conn)
	}
}
