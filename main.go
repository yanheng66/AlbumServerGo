package main

import (
	"database/sql"  // database
	"encoding/json" // JSON
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"         // Gin web framework
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/google/uuid"           // UUID generator
)

// Profile represents the album profile containing artist, title, and year.
type Profile struct {
	Artist string `json:"artist"`
	Title  string `json:"title"`
	Year   string `json:"year"`
}

var db *sql.DB // Global database connection

func main() {
	// Get the database DSN from environment variable DB_DSN
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable is not set")
	}

	// Open a connection to the MySQL database
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}
	defer db.Close()

	// Verify the database connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Error pinging DB: %v", err)
	}

	// Create the albums table if it does not exist
	createTableQuery := `CREATE TABLE IF NOT EXISTS albums (
		album_id VARCHAR(255) PRIMARY KEY,
		image_data LONGBLOB,
		image_size INT NOT NULL,
		artist VARCHAR(255) NOT NULL,
		title VARCHAR(255) NOT NULL,
		year VARCHAR(4) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
	log.Println("Albums table created or already exists.")

	// Clear any existing data in the albums table
	_, err = db.Exec("TRUNCATE TABLE albums;")
	if err != nil {
		log.Fatalf("Error clearing table: %v", err)
	}
	log.Println("Albums table cleared.")

	// Create a Gin router with default middleware (logger and recovery)
	router := gin.Default()

	// Health check endpoint for ALB
	router.GET("/count", func(c *gin.Context) {
		// Return 200 OK for load balancer health check
		c.String(http.StatusOK, "OK")
	})

	// POST /albums endpoint to upload image and profile data, and persist them into the database.
	router.POST("/albums", func(c *gin.Context) {
		// Retrieve the 'image' file from the multipart/form-data request.
		fileHeader, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request: image is required"})
			return
		}

		// Retrieve the 'profile' field as a text string.
		profileStr := c.PostForm("profile")
		if profileStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request: profile is required"})
			return
		}

		// Unmarshal the profile JSON string into a Profile struct.
		var profile Profile
		if err := json.Unmarshal([]byte(profileStr), &profile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request: profile is not valid JSON"})
			return
		}

		// Open the image file.
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "failed to open image file"})
			return
		}
		defer file.Close()

		// Read the image file content.
		imageData, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "failed to read image file"})
			return
		}
		imageSize := int64(len(imageData))

		// Generate a unique albumID.
		albumID := uuid.New().String()

		// Insert the new album record into the database.
		query := `INSERT INTO albums (album_id, image_data, image_size, artist, title, year) VALUES (?, ?, ?, ?, ?, ?)`
		_, err = db.Exec(query, albumID, imageData, imageSize, profile.Artist, profile.Title, profile.Year)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "failed to persist album data"})
			return
		}

		// Return JSON response with albumID and imageSize.
		c.JSON(http.StatusOK, gin.H{
			"albumID":   albumID,
			"imageSize": strconv.FormatInt(imageSize, 10),
		})
	})

	// GET /albums/:albumID endpoint to retrieve album information from the database.
	router.GET("/albums/:albumID", func(c *gin.Context) {
		albumID := c.Param("albumID")
		if albumID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request: albumID is required"})
			return
		}

		// Query the album information from the database.
		var artist, title, year string
		query := `SELECT artist, title, year FROM albums WHERE album_id = ?`
		err := db.QueryRow(query, albumID).Scan(&artist, &title, &year)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"msg": "album not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "failed to retrieve album data"})
			return
		}

		// Return the album information.
		c.JSON(http.StatusOK, gin.H{
			"artist": artist,
			"title":  title,
			"year":   year,
		})
	})

	// Start the server on port 8080
	// Note: Port 8080 is used to match the ALB target group health check configuration.
	router.Run(":8080")
}
