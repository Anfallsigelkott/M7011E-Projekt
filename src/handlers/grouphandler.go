package handlers

import (
	"m7011e-projekt/src/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func FetchGroups(c *gin.Context, db database.Forum_db) {
	groups, err := db.GetGroups()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	}

	c.IndentedJSON(http.StatusOK, groups)
}

func FetchGroupUsers(c *gin.Context, db database.Forum_db) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	}

	users, err := db.GetUsersInGroup(int(id))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	}

	c.IndentedJSON(http.StatusOK, users)
}

func FetchUsersGroups(c *gin.Context, db database.Forum_db) {
	groups, err := db.GetJoinedGroups(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	}

	c.IndentedJSON(http.StatusOK, groups)
}
