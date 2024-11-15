package routes

import (
	"m7011e-projekt/src/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Router(forum_db database.Forum_db) {
	engin := gin.Default()
	engin.Use(cors.Default()) //TODO: Setup better allow-origin policy when more concrete frontend domain is set-up

	//router := engin.Group("/v1")

	// put routes here

	err := engin.Run("localhost:8080")
	if err != nil {
		return
	}
}
