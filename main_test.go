package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupTestDB() *sql.DB {
	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`
		CREATE TABLE cache (
			id VARCHAR(36) PRIMARY KEY,
			content TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		panic(err)
	}
	return db
}

func TestAddHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db = setupTestDB()
	r := gin.Default()
	r.POST("/add", addHandler)

	// Test data
	testContent := "test content"
	reqBody, _ := json.Marshal(map[string]string{"content": testContent})
	req, _ := http.NewRequest("POST", "/add", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["id"])

	// Verify in DB
	var content string
	err := db.QueryRow("SELECT content FROM cache WHERE id = ?", response["id"]).Scan(&content)
	assert.NoError(t, err)
	assert.Equal(t, testContent, content)
}

func TestFetchHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db = setupTestDB()
	r := gin.Default()
	r.GET("/fetch/:id", fetchHandler)

	// Insert test data
	testID := uuid.New().String()
	testContent := "test content"
	_, err := db.Exec("INSERT INTO cache (id, content) VALUES (?, ?)", testID, testContent)
	assert.NoError(t, err)

	// Test request
	req, _ := http.NewRequest("GET", "/fetch/"+testID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, testContent, response["content"])
}

func TestFetchNotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db = setupTestDB()
	r := gin.Default()
	r.GET("/fetch/:id", fetchHandler)

	// Test request with non-existent ID
	req, _ := http.NewRequest("GET", "/fetch/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCleanupOldEntries(t *testing.T) {
	// Setup
	db = setupTestDB()

	// Insert old and new entries
	oldID := uuid.New().String()
	newID := uuid.New().String()
	_, err := db.Exec("INSERT INTO cache (id, content, created_at) VALUES (?, ?, DATETIME('now', '-25 hours'))", oldID, "old content")
	assert.NoError(t, err)
	_, err = db.Exec("INSERT INTO cache (id, content) VALUES (?, ?)", newID, "new content")
	assert.NoError(t, err)

	// Run cleanup
	cleanupOldEntries()

	// Verify old entry is deleted, new entry remains
	var oldContent, newContent string
	err = db.QueryRow("SELECT content FROM cache WHERE id = ?", oldID).Scan(&oldContent)
	assert.Error(t, err) // Should not exist
	err = db.QueryRow("SELECT content FROM cache WHERE id = ?", newID).Scan(&newContent)
	assert.NoError(t, err)
	assert.Equal(t, "new content", newContent)
}
