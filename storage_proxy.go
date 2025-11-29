package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

func RunStorageProxy() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, os.ModePerm)
	}

	r.POST("/upload", func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File tidak ditemukan"})
			return
		}
		defer file.Close()

		ext := filepath.Ext(header.Filename)
		var uuidName string
		var filePath string

		for {
			uuidName = uuid.New().String() + ext
			filePath = filepath.Join(uploadDir, uuidName)

			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				break // Exit loop if file already exists
			}
		}

		out, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file"})
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file"})
			return
		}

		url := fmt.Sprintf("/file/%s", uuidName)
		c.JSON(http.StatusOK, gin.H{"fileUrl": url})
	})

	r.GET("/file/:filename", func(c *gin.Context) {
		fmt.Println("filename", c.Param("filename"))
		filename := c.Param("filename")
		filePath := filepath.Join(uploadDir, filename)
		c.File(filePath)
	})

	r.Run(":8081")
}
