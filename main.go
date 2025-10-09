package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var db *sql.DB

db, err = sql.Open("duckdb", "./cache.db")


func main() {
	// Initialize DuckDB
	var err error
	db, err = sql.Open("duckdb", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create cache table
	_, err = db.Exec(`
		CREATE TABLE cache (
			id VARCHAR(36) PRIMARY KEY,
			content TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Start cleanup goroutine
	go cleanupOldEntries()

	// Set up Gin router
	r := gin.Default()
	r.POST("/add", addHandler)
	r.GET("/fetch/:id", fetchHandler)
	r.Run(":8080")
}

func addHandler(c *gin.Context) {
	var request struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	_, err := db.Exec("INSERT INTO cache (id, content) VALUES (?, ?)", id, request.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func fetchHandler(c *gin.Context) {
	id := c.Param("id")
	var content string
	err := db.QueryRow("SELECT content FROM cache WHERE id = ?", id).Scan(&content)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"content": content})
}

func cleanupOldEntries() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		_, err := db.Exec("DELETE FROM cache WHERE created_at < DATETIME('now', '-24 hours')")
		if err != nil {
			log.Printf("Cleanup error: %v", err)
		}
	}
}
