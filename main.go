package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Keys struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func main() {
	// etcd client initialization
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://etcd-container:23790"},
	})
	if err != nil {
		panic(err)
	}
	defer etcdClient.Close()

	// Gin router initialization
	router := gin.Default()

	router.GET("/:key", func(c *gin.Context) {
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
	})

	router.GET("", func(c *gin.Context) {
		resp, err := etcdClient.Get(c.Request.Context(), "", clientv3.WithPrefix())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var keys []Keys
		for _, kv := range resp.Kvs {
			key := string(kv.Key) // Assuming the key is stored as a raw string
			value := string(kv.Value)
			keys = append(keys, Keys{Key: key, Value: value})
		}

		c.JSON(http.StatusOK, gin.H{"keys": keys})
	})

	router.POST("", func(c *gin.Context) {
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
	})

	// Run the server
	router.Run(":8080")
}
