package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"m7011e-projekt/database"
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
	if err != nil { // we expect error here if the user isn't in the group (no valid row)
		fmt.Printf(err.Error())
		c.IndentedJSON(http.StatusForbidden, err.Error())
		return
	}

	replyID := int64(0)
	if len(body["replyID"]) > 0 {
		replyID, err = strconv.ParseInt(body["replyID"], 10, 64)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err.Error())
			return
		}
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

	userEntry, err := db.GetUserByUsername(user)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	if (role != 2 && !userEntry.IsAdmin) || len(body["content"]) > 0 { // Check only necessary if user isn't an administrator/mod, if admin bypass then new content must be empty for delete
		_, err = db.MatchUserToPost(user, int(postID))
		if err != nil { // we expect error here if the user did not create the relevant post (no valid row)
			c.IndentedJSON(http.StatusForbidden, err.Error())
			return
		}
	}
	fmt.Printf("content: %v", body["content"])
	content := body["content"]
	err = db.UpdatePostContent(int(postID), content)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, nil)
}

func FetchPosts(c *gin.Context, db database.Forum_db) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	}
	posts, err := db.GetPostsInGroup(int(id))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	}
	c.IndentedJSON(http.StatusOK, posts)
}
