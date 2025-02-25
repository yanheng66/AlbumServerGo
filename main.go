package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Profile represents the album profile containing artist, title, and year.
type Profile struct {
	Artist string `json:"artist"`
	Title  string `json:"title"`
	Year   string `json:"year"`
}

func main() {
	// Create a Gin router with default middleware (logger and recovery)
	router := gin.Default()

	// POST /albums endpoint to handle file upload and profile data
	router.POST("/albums", func(c *gin.Context) {
		// Retrieve the 'image' file from the multipart/form-data request.
		file, err := c.FormFile("image")
		if err != nil {
			// Return HTTP 400 if the image part is missing (HTTP 400: Bad Request)
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

		// For this stub, return a fixed albumID and the image file size.
		fixedAlbumID := "fixedAlbumId"
		imageSize := strconv.FormatInt(file.Size, 10) // Convert file size (int64) to string

		// Return JSON response with albumID and imageSize.
		c.JSON(http.StatusOK, gin.H{
			"albumID":   fixedAlbumID,
			"imageSize": imageSize,
		})
	})

	// GET /albums/:albumID endpoint to retrieve album information based on albumID.
	router.GET("/albums/:albumID", func(c *gin.Context) {
		// Retrieve albumID from the URL path parameter.
		albumID := c.Param("albumID")
		if albumID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "invalid request: albumID is required"})
			return
		}

		// For this stub, always return constant album information.
		c.JSON(http.StatusOK, gin.H{
			"artist": "Sex Pistols",
			"title":  "Never Mind The Bollocks!",
			"year":   "1977",
		})
	})

	// Start the server on port 8080.
	router.Run(":8081")
}
