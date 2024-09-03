package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/redis.v5"
)

// Item represents a basic item in our CRUD application
type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var (
	items = make(map[string]Item)
)

// Logger is a globally accessible instance of logrus.Logger
var Logger *logrus.Logger

// Redis client
var client *redis.Client

func init() {
	// Initialize Redis client
	client = redis.NewClient(&redis.Options{
		Addr: "fedora:6379",
	})

	_, err := client.Ping().Result()
	if err != nil {
		fmt.Println("Could not connect to Redis:", err)
		os.Exit(1)
	}

	// Initialize the logger
	Logger = logrus.New()
	Logger.SetFormatter(&logrus.JSONFormatter{})

	// Send logs to Logstash
	Logger.Out = os.Stdout
	logstashHook, err := NewLogstashHook("tcp", "fedora:5000", "golang-crud-app")
	if err != nil {
		fmt.Println("Could not create Logstash hook:", err)
		os.Exit(1)
	}
	Logger.Hooks.Add(logstashHook)
}

func main() {
	router := gin.Default()

	// CRUD routes
	router.POST("/items", createItem)
	router.GET("/items", getItems)
	router.GET("/items/:id", getItem)
	router.PUT("/items/:id", updateItem)
	router.DELETE("/items/:id", deleteItem)

	// Start the server
	port := ":3000"
	Logger.WithField("port", port).Info("Starting server")
	router.Run(port)
}

// createItem handles the creation of a new item
func createItem(c *gin.Context) {
	var newItem Item
	if err := c.ShouldBindJSON(&newItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newItem.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	items[newItem.ID] = newItem

	Logger.WithFields(logrus.Fields{
		"action":      "create",
		"item_id":     newItem.ID,
		"item_name":   newItem.Name,
		"description": newItem.Description,
	}).Info("Item created")

	// Store the item in Redis
	itemJSON, _ := json.Marshal(newItem)
	client.Set(newItem.ID, itemJSON, 0)

	c.JSON(http.StatusCreated, newItem)
}

// getItems retrieves all items
func getItems(c *gin.Context) {
	var itemList []Item
	for _, item := range items {
		itemList = append(itemList, item)
	}
	c.JSON(http.StatusOK, itemList)
}

// getItem retrieves a single item by ID
func getItem(c *gin.Context) {
	id := c.Param("id")
	item, exists := items[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"message": "Item not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// updateItem updates an existing item
func updateItem(c *gin.Context) {
	id := c.Param("id")
	var updatedItem Item
	if err := c.ShouldBindJSON(&updatedItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, exists := items[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"message": "Item not found"})
		return
	}

	item.Name = updatedItem.Name
	item.Description = updatedItem.Description
	items[id] = item

	Logger.WithFields(logrus.Fields{
		"action":      "update",
		"item_id":     id,
		"item_name":   item.Name,
		"description": item.Description,
	}).Info("Item updated")

	// Update the item in Redis
	itemJSON, _ := json.Marshal(item)
	client.Set(id, itemJSON, 0)

	c.JSON(http.StatusOK, item)
}

// deleteItem deletes an item by ID
func deleteItem(c *gin.Context) {
	id := c.Param("id")
	_, exists := items[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"message": "Item not found"})
		return
	}

	delete(items, id)
	Logger.WithFields(logrus.Fields{
		"action":  "delete",
		"item_id": id,
	}).Info("Item deleted")

	// Remove the item from Redis
	client.Del(id)

	c.JSON(http.StatusOK, gin.H{"message": "Item deleted"})
}

// NewLogstashHook creates a Logstash hook for logrus
func NewLogstashHook(protocol, address, appName string) (logrus.Hook, error) {
	conn, err := net.Dial(protocol, address)
	if err != nil {
		return nil, err
	}
	hook := &LogstashHook{
		writer: conn,
		appName: appName,
	}
	return hook, nil
}

// LogstashHook is a custom hook for sending logs to Logstash
type LogstashHook struct {
	writer  net.Conn
	appName string
}

// Levels returns the levels supported by this hook
func (hook *LogstashHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire sends the log entry to Logstash
func (hook *LogstashHook) Fire(entry *logrus.Entry) error {
	entry.Data["app"] = hook.appName
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.writer.Write([]byte(line))
	return err
}
