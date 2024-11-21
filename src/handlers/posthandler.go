package handlers

import (
	"encoding/json"
	"io"
	"m7011e-projekt/src/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreatePost(c *gin.Context, db database.Forum_db) {
	bodyAsByteArray, _ := io.ReadAll(c.Request.Body)
	body := make(map[string]string)
	json.Unmarshal(bodyAsByteArray, &body)

	group, err := strconv.ParseInt(c.Param("group"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	tokenstring, err := c.Cookie("authtoken")
	user, err := ExtractJWT(tokenstring)
	_, err = db.GetRoleInGroup(int(group), user)
	if err != nil { // we expect error her if the user isn't in the group (no valid row)
		c.IndentedJSON(http.StatusForbidden, err.Error())
		return
	}

	replyID, err := strconv.ParseInt(body["replyID"], 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	err = db.CreatePostEntry(user, int(group), body["content"], int(replyID))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	}
	c.IndentedJSON(http.StatusOK, nil)
}

func UpdatePost(c *gin.Context, db database.Forum_db) { // Delete can be done through this call too
	bodyAsByteArray, _ := io.ReadAll(c.Request.Body)
	body := make(map[string]string)
	json.Unmarshal(bodyAsByteArray, &body)

	group, err := strconv.ParseInt(c.Param("group"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	tokenstring, err := c.Cookie("authtoken")
	user, err := ExtractJWT(tokenstring)
	role, err := db.GetRoleInGroup(int(group), user)
	if err != nil { // we expect error here if the user isn't in the group (no valid row)
		c.IndentedJSON(http.StatusForbidden, err.Error())
		return
	}

	postID, err := strconv.ParseInt(body["postID"], 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	if role != 3 || len(body["content"]) > 0 { // Check only necessary if user isn't an administrator, if admin bypass then new content must be empty for delete
		_, err = db.MatchUserToPost(user, int(postID))
		if err != nil { // we expect error here if the user did not create the relevant post (no valid row)
			c.IndentedJSON(http.StatusForbidden, err.Error())
			return
		}
	}
	err = db.UpdatePostContent(int(postID), body["content"])
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	}
	c.IndentedJSON(http.StatusOK, nil)
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
