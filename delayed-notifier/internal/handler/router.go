package handler

import (
	"github.com/wb-go/wbf/ginext"
)

func NewRouter(notifyHandler *NotifyHandler) *ginext.Engine {
	router := ginext.New("release")
	router.Use(MetricsMiddleware)
	router.Use(ginext.Logger())
	router.Use(ginext.Recovery())
	router.StaticFile("/", "/app/internal/static/index.html")
	router.POST("/notify", notifyHandler.CreateNotification)
	router.GET("/notify", notifyHandler.GetAllNotifications)
	router.GET("/notify/:id", notifyHandler.GetNotification)
	router.DELETE("/notify/:id", notifyHandler.DeleteNotification)
	router.GET("/metrics", notifyHandler.Metrics)
	return router
}
