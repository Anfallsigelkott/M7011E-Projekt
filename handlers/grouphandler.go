package handlers

import (
	"m7011e-projekt/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateNewGroup(c *gin.Context, db database.Forum_db) {
	groupname := c.Param("group")
	err := db.CreateNewGroup(groupname)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, nil)
}

func DeleteGroup(c *gin.Context, db database.Forum_db) { // Needs admin-specific middleware
	groupID, err := strconv.ParseInt(c.Param("group"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	err = db.RemoveGroup(int(groupID))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, nil)
}

func JoinGroup(c *gin.Context, db database.Forum_db) {
	groupID, err := strconv.ParseInt(c.Param("group"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	username := c.Param("user")
	err = db.AddUserToGroup(username, int(groupID), 1) // default to setting new joinees as regular users?
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, nil)
}

func RemoveFromGroup(c *gin.Context, db database.Forum_db) {
	groupID, err := strconv.ParseInt(c.Param("group"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	username := c.Param("user")
	_, err = db.GetRoleInGroup(int(groupID), username)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, "no such user found in the group")
		return
	}

	tokenstring, err := c.Cookie("authtoken")
	actinguser, err := ExtractJWT(tokenstring)
	actingrole, err := db.GetRoleInGroup(int(groupID), actinguser)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}

	actingUserEntry, err := db.GetUserByUsername(actinguser)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	if actingrole != 2 && !actingUserEntry.IsAdmin && actinguser != username { // If user submitting leave request isn't admin or the leaving user, reject request
		c.IndentedJSON(http.StatusForbidden, "Removal request submitted for non-self user by non-admin")
		return
	}
	err = db.RemoveUserFromGroup(username, int(groupID))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, nil)
}

func UpdateUserRoleInGroup(c *gin.Context, db database.Forum_db) {
	groupID, err := strconv.ParseInt(c.Param("groupID"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	newRole, err := strconv.ParseInt(c.Param("newRole"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	if newRole > 2 {
		newRole = 2
	} else if newRole < 1 {
		newRole = 1
	} // ensuring our roles remain within specified bounds

	tokenstring, err := c.Cookie("authtoken")
	user, err := ExtractJWT(tokenstring)
	role, err := db.GetRoleInGroup(int(groupID), user)
	if err != nil { // we expect error here if the user isn't in the group (no valid row)
		c.IndentedJSON(http.StatusForbidden, err.Error())
		return
	}
	userEntry, err := db.GetUserByUsername(user)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	if role != 2 && !userEntry.IsAdmin {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "must be group moderator or admin to update role"})
		return
	}

	user = c.Param("user")
	err = db.UpdateUserRole(user, int(groupID), int(newRole))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, nil)
}

func FetchGroups(c *gin.Context, db database.Forum_db) {
	groups, err := db.GetGroups()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, groups)
}

func FetchGroupUsers(c *gin.Context, db database.Forum_db) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	users, err := db.GetUsersInGroup(int(id))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, users)
}

func FetchUsersGroups(c *gin.Context, db database.Forum_db) {
	groups, err := db.GetJoinedGroups(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, groups)
}
