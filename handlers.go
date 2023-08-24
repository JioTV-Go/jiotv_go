package main

import (
	"github.com/gin-gonic/gin"
)

func indexHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Success",
	})
}

func loginHandler(c *gin.Context) {
	username, check := c.GetQuery("username")
	password, check := c.GetQuery("password")
	if !check {
		Log.Println("Username or password not provided")	
		c.JSON(400, gin.H{
			"message": "Username or password not provided",
		})
		return
	}
	// Login
	result, err := Login(username, password)
	if err != nil {
		Log.Println(err)
		return
	}
	c.JSON(200, result)

}

func getLive(c *gin.Context) {
	id := c.Param("id")
	result, err := GetLive(id)
	if err != nil {
		Log.Println(err)
		return
	}
	c.JSON(200, result)
}
