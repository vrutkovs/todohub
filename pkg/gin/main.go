package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/contrib/static"
	"net/http"
)

func SetupGin() *gin.Engine {
	r := gin.New()

	// Server static HTML
	r.Use(static.Serve("/", static.LocalFile("./html", true)))

	// Don't log k8s health endpoint
	r.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/health"),
		gin.Recovery(),
	)

	setupRoutes(r)

	return r
}

func StartGin(r *gin.Engine) {
	r.Run(":8080")
}

func setupRoutes(r *gin.Engine) {
	// Generic webhook, which would display output in markdown
	r.GET("/health", health)
}

func health(c *gin.Context) {
	c.String(http.StatusOK, "")
}
