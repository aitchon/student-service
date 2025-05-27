package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"student-service/config"
	"student-service/models"
	"student-service/repositories"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Load test configuration
	cfg, err := config.LoadConfig("config/config_test.yaml")
	if err != nil {
		panic(err)
	}

	// Initialize database with test database
	db, err := sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		panic(err)
	}

	// Create test table
	createTableSQL := `
CREATE TABLE IF NOT EXISTS students (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		age INTEGER,
		grade TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		panic(err)
	}

	// Initialize repository and router
	studentRepo := repositories.NewStudentRepository(db)
	r := gin.Default()
	setupRoutes(r, studentRepo)

	return r
}

func TestGetAllStudents(t *testing.T) {
	router := setupTestRouter()

	// Test empty list
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/students", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var response []models.Student
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Empty(t, response)
}

func TestCreateStudent(t *testing.T) {
	router := setupTestRouter()

	// Test successful creation
	student := models.Student{
		Name:  "John Doe",
		Grade: 98,
	}
	jsonData, _ := json.Marshal(student)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/students", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
	var response models.Student
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, student.Name, response.Name)
	assert.Equal(t, student.Grade, response.Grade)
	assert.NotZero(t, response.ID)
}

func TestGetStudentByID(t *testing.T) {
	router := setupTestRouter()

	// First create a student
	student := models.Student{
		Name:  "Jane Doe",
		Grade: 88,
	}
	jsonData, _ := json.Marshal(student)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/students", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var createdStudent models.Student
	json.Unmarshal(w.Body.Bytes(), &createdStudent)

	// Test getting the student
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/students/"+strconv.Itoa(createdStudent.ID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var response models.Student
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, createdStudent.ID, response.ID)
	assert.Equal(t, student.Name, response.Name)
}

func TestUpdateStudent(t *testing.T) {
	router := setupTestRouter()

	// First create a student
	student := models.Student{
		Name:  "Bob Smith",
		Grade: 75,
	}
	jsonData, _ := json.Marshal(student)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/students", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var createdStudent models.Student
	json.Unmarshal(w.Body.Bytes(), &createdStudent)

	// Test updating the student
	updatedStudent := models.Student{
		Name:  "Bob Smith Updated",
		Grade: 83,
	}
	jsonData, _ = json.Marshal(updatedStudent)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/students/"+strconv.Itoa(createdStudent.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var response models.Student
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, updatedStudent.Name, response.Name)
	assert.Equal(t, updatedStudent.Grade, response.Grade)
}

func TestDeleteStudent(t *testing.T) {
	router := setupTestRouter()

	// First create a student
	student := models.Student{
		Name:  "Alice Johnson",
		Grade: 100,
	}
	jsonData, _ := json.Marshal(student)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/students", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var createdStudent models.Student
	json.Unmarshal(w.Body.Bytes(), &createdStudent)

	// Test deleting the student
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/students/"+strconv.Itoa(createdStudent.ID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 204, w.Code)

	// Verify student is deleted
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/students/"+strconv.Itoa(createdStudent.ID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}
