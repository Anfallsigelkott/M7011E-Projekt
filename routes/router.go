package routes

import (
	"m7011e-projekt/database"
	"m7011e-projekt/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Router(forum_db database.Forum_db) {
	engin := gin.Default()
	engin.Use(cors.Default()) //TODO: Setup better allow-origin policy when more concrete frontend domain is set-up

	router := engin.Group("/v1")

	// Getters

	router.GET("/groups", func(ctx *gin.Context) { handlers.FetchGroups(ctx, forum_db) }) //GET groups
	//router.GET("/user/:id", func(ctx *gin.Context) {})                                            //GET UserByID
	router.GET("/groups/user/:id", func(ctx *gin.Context) { handlers.FetchUsersGroups(ctx, forum_db) }) //GET JoinedGroups
	router.GET("/groups/:id/user", func(ctx *gin.Context) { handlers.FetchGroupUsers(ctx, forum_db) })  //GET UsersInGroup
	router.GET("/groups/:id/post", func(ctx *gin.Context) { handlers.FetchPosts(ctx, forum_db) })       //GET PostsInGroup
	//router.GET("/groups/:id/role", func(ctx *gin.Context) {}) //GET RoleInGroup

	// Creators

	router.POST("/user/new", func(ctx *gin.Context) { handlers.RegisterUser(ctx, forum_db) })             //POST CreateNewUser
	router.POST("/group/new/:group", func(ctx *gin.Context) { handlers.CreateNewGroup(ctx, forum_db) })   //POST CreateNewGroup
	router.POST("/post/:group", func(ctx *gin.Context) { handlers.CreatePost(ctx, forum_db) })            //POST CreatePostEntry
	router.POST("/user/:user/join/:group/", func(ctx *gin.Context) { handlers.JoinGroup(ctx, forum_db) }) //POST AddUserToGroup

	// Updaters

	router.POST("/user/:user/role/:groupID/:newRole", func(ctx *gin.Context) { handlers.UpdateUserRoleInGroup(ctx, forum_db) }) //POST UpdateUserRole
	router.POST("/post/:group/edit", func(ctx *gin.Context) { handlers.UpdatePost(ctx, forum_db) })                             //POST UpdatePostContent
	router.POST("/user/edit", func(ctx *gin.Context) { handlers.UpdateUsername(ctx, forum_db) })                                //POST UpdateUsername (and mby other user option thingymayigs)

	// Removers

	router.POST("/user/:user/leave/:group", func(ctx *gin.Context) { handlers.RemoveFromGroup(ctx, forum_db) })                                                              //POST RemoveUserFromGroup
	router.POST("/group/:group/delete", func(ctx *gin.Context) { handlers.AdminValidateJWT(ctx, forum_db) }, func(ctx *gin.Context) { handlers.DeleteGroup(ctx, forum_db) }) //POST RemoveGroup

	router.POST("/user/login", func(ctx *gin.Context) { handlers.LoginUser(ctx, forum_db) }) //Bloody login init??

	err := engin.Run("0.0.0.0:8080")
	if err != nil {
		return
	}
}
