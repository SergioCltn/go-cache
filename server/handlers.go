package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

func (cs *CacheServer) handleGetAll(conn net.Conn) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	var allItems []CacheEntry

	for _, item := range cs.cache {
		item.mu.Lock()
		allItems = append(allItems, item)
		item.mu.Unlock()
	}

	result, _ := json.Marshal(allItems)

	conn.Write(append(result, '\n'))

}

func (cs *CacheServer) handleGet(conn net.Conn, key string) {
	cs.mu.Lock()
	valueCache, ok := cs.cache[key]
	cs.mu.Unlock()

	if !ok {
		conn.Write([]byte("This key is not in the cache\n"))
		return
	}
	result, _ := json.Marshal(valueCache)
	conn.Write(append(result, '\n'))

}
func (cs *CacheServer) handleSet(conn net.Conn, key string, value string, ttl string) {
	parsedTTL, err := strconv.Atoi(ttl)
	if err != nil {
		conn.Write([]byte("Invalid TTL\n"))
		return
	}
	cs.mu.Lock()

	item, ok := cs.cache[key]
	if !ok {
		item = CacheEntry{Key: key, mu: &sync.Mutex{}}
	}
	item.Value = value
	item.TTL = time.Now().Add(time.Duration(parsedTTL) * time.Minute)
	cs.cache[key] = item

	cs.mu.Unlock()

	conn.Write([]byte("Okkkk manin\n"))
}

func (cs *CacheServer) handleDelete(conn net.Conn, key string) {

	cs.mu.Lock()

	_, ok := cs.cache[key]
	if !ok {
		conn.Write([]byte("Manin que esta clave no existe que dise\n"))
		return
	}
	delete(cs.cache, key)

	cs.mu.Unlock()

	conn.Write([]byte(fmt.Sprintf("Ya te he borrado la key: %s, gracias fenomeno!\n", key)))
}
