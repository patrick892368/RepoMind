package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()
	router.GET("/wallet/info", walletInfo)
}

func walletInfo(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"ok": true})
}
