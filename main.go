package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"student-service/config"
	"student-service/models"
	"student-service/repositories"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create students table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS students (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		age INTEGER,
		grade TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize repository
	studentRepo := repositories.NewStudentRepository(db)

	// Initialize router
	r := gin.Default()

	// Configure CORS
	// had to add CORS middleware to allow requests from localhost:3000 to debug locally
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     cfg.CORS.AllowedMethods,
		AllowHeaders:     cfg.CORS.AllowedHeaders,
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	setupRoutes(r, studentRepo)

	// Start the server with configured port
	serverAddr := ":" + strconv.Itoa(cfg.Server.Port)
	if err := r.Run(serverAddr); err != nil {
		log.Fatal(err)
	}
}

func setupRoutes(r *gin.Engine, studentRepo *repositories.StudentRepository) {
	// GET /students - List all students or all students with a specific name
	r.GET("/students", func(c *gin.Context) {
		name := c.Query("name")

		students, err := studentRepo.GetAllStudents(name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, students)
	})

	// GET /students/:id - Get a specific student
	r.GET("/students/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
			return
		}

		student, err := studentRepo.GetStudentByID(id)
		if err != nil {
			if err == repositories.ErrStudentNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, student)
	})

	// POST /students - Create a new student
	r.POST("/students", func(c *gin.Context) {
		var student models.Student
		if err := c.ShouldBindJSON(&student); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := studentRepo.CreateStudent(&student)
		if err != nil {
			if err == repositories.ErrDuplicateName {
				c.JSON(http.StatusConflict, gin.H{"error": "student with this name already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, student)
	})

	// PUT /students/:id - Update a student
	r.PUT("/students/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
			return
		}

		var student models.Student
		if err := c.ShouldBindJSON(&student); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		student.ID = id
		err = studentRepo.UpdateStudent(&student)
		if err != nil {
			if err == repositories.ErrStudentNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, student)
	})

	// DELETE /students/:id - Delete a student
	r.DELETE("/students/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
			return
		}

		err = studentRepo.DeleteStudent(id)
		if err != nil {
			if err == repositories.ErrStudentNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	})
}
