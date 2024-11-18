package handlers

import (
	"m7011e-projekt/src/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreatePost(c *gin.Context, db database.Forum_db) {

}

func FetchPosts(c *gin.Context, db database.Forum_db) {
	id, err := strconv.ParseInt(c.Param("group"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	}
	posts, err := db.GetPostsInGroup(int(id))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	}
	c.IndentedJSON(http.StatusOK, posts)
}
