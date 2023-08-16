package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Keys struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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

	wg.Add(1)
	go watchRequests()

	router := gin.Default()

	router.GET("/:key", handleGetKey)
	router.GET("", handleListKeys)
	router.POST("", handleAddKey)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		logger.Println("Shutting down gracefully...")
		wg.Wait()
		os.Exit(0)
	}()

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

	file, err := os.OpenFile("output.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Printf("Error opening file: %v", err)
		return
	}
	defer file.Close()

	str := fmt.Sprintf("Change detected for key %s: value:%s Type=%s\n", string(event.Kv.Key), string(event.Kv.Value), event.Type)
	file.WriteString(str)
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
