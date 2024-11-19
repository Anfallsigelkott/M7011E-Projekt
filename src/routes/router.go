package routes

import (
	"m7011e-projekt/src/database"
	"m7011e-projekt/src/handlers"

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

	router.POST("/user/new", func(ctx *gin.Context) { handlers.RegisterUser(ctx, forum_db) }) //POST CreateNewUser
	router.POST("/group/new", func(ctx *gin.Context) {})                                      //POST CreateNewGroup
	router.POST("/post", func(ctx *gin.Context) { handlers.CreatePost(ctx, forum_db) })       //POST CreatePostEntry
	router.POST("/user/join", func(ctx *gin.Context) {})                                      //POST AddUserToGroup

	// Updaters

	router.POST("/user/role", func(ctx *gin.Context) {}) //POST UpdateUserRole
	router.POST("/post/edit", func(ctx *gin.Context) {}) //POST UpdatePostContent
	router.POST("/user/edit", func(ctx *gin.Context) {}) //POST UpdateUsername (and mby other user option thingymayigs)

	// Removers

	router.POST("/user/leave", func(ctx *gin.Context) {})   //POST RemoveUserFromGroup
	router.POST("/group/delete", func(ctx *gin.Context) {}) //POST RemoveGroup

	router.POST("/user/login", func(ctx *gin.Context) { handlers.LoginUser(ctx, forum_db) }) //Bloody login init??

	err := engin.Run("localhost:8080")
	if err != nil {
		return
	}
}
