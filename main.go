package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Keys struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func main() {
	// Initialize etcd client and watch Goroutine
	initEtcdClient()

	// Initialize Gin router
	router := gin.Default()

	// Route for handling individual keys
	router.GET("/:key", handleGetKey)

	// Route for listing all keys
	router.GET("", handleListKeys)

	// Route for adding a key-value pair
	router.POST("", handleAddKey)

	// Run the server
	router.Run(":8080")
}

func initEtcdClient() {
	var err error
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints: []string{"http://etcd-container:23790"},
	})
	if err != nil {
		panic(err)
	}

	// Create a channel to pass watch events
	eventChan := make(chan *clientv3.Event)

	// Start the watchKeyChanges Goroutine
	go watchKeyChanges(etcdClient, "key-", eventChan)

	// Start the logChange Goroutine
	go func() {
		for event := range eventChan {
			key := string(event.Kv.Key)
			if strings.HasPrefix(key, "key-") {
				logChange(key, event.Type.String())
			}
		}
	}()
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

func logChange(key string, eventType string) {
	logLine := fmt.Sprintf("Change detected for key %s: Type=%s\n", key, eventType)

	logFilePath := "/logfile.txt" // Replace with the actual path
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %s\n", err)
		return
	}
	defer logFile.Close()

	if _, err := logFile.WriteString(logLine); err != nil {
		fmt.Printf("Error writing to log file: %s\n", err)
	}
}

var etcdClient *clientv3.Client

func watchKeyChanges(etcdClient *clientv3.Client, keyPrefixToWatch string, eventChan chan<- *clientv3.Event) {
	watchCtx, watchCancel := context.WithCancel(context.Background())
	defer watchCancel()

	watcher := clientv3.NewWatcher(etcdClient)

	watchChan := watcher.Watch(watchCtx, keyPrefixToWatch, clientv3.WithPrefix())

	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			eventChan <- event
		}
	}
}
func startRedirector() {
	// Create a reverse proxy for etcd
	etcdProxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = "127.0.0.1:2379" // Use the correct IP and port for your etcd instance
			req.Host = req.URL.Host
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Forward requests to etcd instance
		etcdProxy.ServeHTTP(w, r)
	})

	// Start the HTTP server
	http.ListenAndServe(":23791", nil)
}
