package router

import "github.com/gin-gonic/gin"

func Router() *gin.Engine {
	router := gin.Default()

	// ==========================
	// should route web page here
	// ==========================

	// non-json here

	router.Use(func(c *gin.Context) {
		c.SetAccepted("application/json")
		c.Next()
	})
	routeAuth(router.Group("/auth"))
	routeUsers(router.Group("/users"))
	routePosts(router.Group("/posts"))

	// 404 here

	return router
}
