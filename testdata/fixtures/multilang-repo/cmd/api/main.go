package main

import "github.com/gin-gonic/gin"

func main() {
    router := gin.Default()
    router.POST("/login", loginHandler)
    router.GET("/wallet/info", walletInfo)
}

func loginHandler(c *gin.Context) {}
func walletInfo(c *gin.Context) {}
