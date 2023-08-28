package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "time/tzdata"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Keys represents a key-value pair.
type Keys struct {
	Key   string `toml:"key" json:"key"`
	Value string `toml:"value" json:"value"`
}

// Config represents the configuration structure for events.
type Config struct {
	Event struct {
		Key   string `toml:"key"`
		Value string `toml:"value"`
		Time  string `toml:"time"`
	}
}

var (
	etcdClient *clientv3.Client
	logger     *log.Logger
	eventLock  sync.Mutex
	watcher    clientv3.Watcher
	wg         sync.WaitGroup
)

func main() {
	initLogger()     // Initialize the logger.
	initEtcdClient() // Initialize the etcd client.

	watcher = etcdClient.Watcher

	go watchRequests() // Start watching for etcd events in a separate goroutine.

	router := gin.Default() // Initialize the Gin router.

	// Define endpoints and their corresponding handlers.
	router.GET("/:key", handleGetKey)
	router.GET("", handleListKeys)
	router.POST("", handleAddKey)

	router.Run(":8080") // Start the server on port 8080.
}

// Initializes the logger.
func initLogger() {
	logger = log.New(os.Stdout, "[EVENT] ", log.Ldate|log.Ltime)
}

// Initializes the etcd client.
func initEtcdClient() {
	var err error
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints: []string{"http://etcd1:2379", "http://etcd2:2379", "http://etcd3:2379"},
	})
	if err != nil {
		logger.Fatalf("Failed to initialize etcd client: %v", err)
	}
}

// Watches for etcd events.
func watchRequests() {
	defer wg.Done()

	watchKey := "key"
	watchCtx, watchCancel := context.WithCancel(context.Background())
	defer watchCancel()

	// Start watching for events related to 'watchKey'.
	watchChan := watcher.Watch(watchCtx, watchKey, clientv3.WithPrefix())

	// Process events received on the watch channel.
	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			handleEvent(event) // Handle the event.
		}
	}
}

// Handles an etcd event.
func handleEvent(event *clientv3.Event) {
	eventLock.Lock()
	defer eventLock.Unlock()

	// Load Istanbul timezone.
	loca, err := time.LoadLocation("Turkey")
	if err != nil {
		logger.Printf("could not load Istanbul time zone: %v", err)
		return
	}
	now := time.Now().In(loca).Format(http.TimeFormat)

	var config Config
	config.Event.Key = string(event.Kv.Key)
	config.Event.Value = string(event.Kv.Value)
	config.Event.Time = now

	filePath := "/app/data/events.toml"

	// Open or create the events file.
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Printf("couldnt open file: %v", err)
		return
	}
	defer file.Close()

	// Encode and write the event data to the file.
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		logger.Printf("couldnt encode: %v", err)
	}
}

// Handles the retrieval of a specific key's value.
func handleGetKey(c *gin.Context) {
	key := c.Param("key")
	resp, err := etcdClient.Get(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(resp.Kvs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Key not found"})
		return
	}
	value := string(resp.Kvs[0].Value)
	c.JSON(http.StatusOK, gin.H{"key": key, "value": value})
}

// Handles listing keys and their values.
func handleListKeys(c *gin.Context) {
	resp, err := etcdClient.Get(c.Request.Context(), "", clientv3.WithPrefix())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var keys []Keys
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		value := string(kv.Value)
		keys = append(keys, Keys{Key: key, Value: value})
	}

	c.JSON(http.StatusOK, gin.H{"keys": keys})
}

// Handles adding a new key-value pair.
func handleAddKey(c *gin.Context) {
	var data Keys
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := etcdClient.Put(c.Request.Context(), data.Key, data.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Key-value pair added successfully"})
}
