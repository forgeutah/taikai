package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()
	r.LoadHTMLGlob("views/*")
	r.Static("/assets", "./assets")

	r.GET("/", func(c *gin.Context) {
		content := "hello world"
		c.HTML(http.StatusOK, "index", gin.H{
			"Title":      "Taikai: Meetup for free",
			"IsLoggedIn": false,
			"Content":    content,
		})
	})
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login", gin.H{
			"Username": "",
			"Password": "",
		})
	})

	r.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup", gin.H{})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run()

}
