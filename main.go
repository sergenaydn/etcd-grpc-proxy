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

type Keys struct {
	Key   string `toml:"key" json:"key"`
	Value string `toml:"value" json:"value"`
}

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
	initLogger()

	initEtcdClient()

	watcher = etcdClient.Watcher

	go watchRequests()

	router := gin.Default()

	router.GET("/:key", handleGetKey)
	router.GET("", handleListKeys)
	router.POST("", handleAddKey)
	// Run the server
	router.Run(":8080")
}

func initLogger() {
	logger = log.New(os.Stdout, "[EVENT] ", log.Ldate|log.Ltime)
}

func initEtcdClient() {
	var err error
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints: []string{"http://etcd1:2379", "http://etcd2:2379", "http://etcd3:2379"},
	})
	if err != nil {
		logger.Fatalf("Failed to initialize etcd client: %v", err)
	}
}

func watchRequests() {
	defer wg.Done()

	watchKey := "key"
	watchCtx, watchCancel := context.WithCancel(context.Background())
	defer watchCancel()

	watchChan := watcher.Watch(watchCtx, watchKey, clientv3.WithPrefix())

	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			handleEvent(event)
		}
	}
}

func handleEvent(event *clientv3.Event) {
	eventLock.Lock()
	defer eventLock.Unlock()

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

	// Use the correct path within the mounted volume
	filePath := "/app/data/events.toml"

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Printf("couldnt open file: %v", err)
		return
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		logger.Printf("couldnt encode: %v", err)
	}
}

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
