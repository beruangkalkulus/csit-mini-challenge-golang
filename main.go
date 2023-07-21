package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/", indexHandler)

	router.Run("localhost:8080")
}

func indexHandler(c *gin.Context) {
	fmt.Println("indexHandler")
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello World!",
	})
}