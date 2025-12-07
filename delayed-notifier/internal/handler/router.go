package handler

import "github.com/wb-go/wbf/ginext"

func NewRouter(notifyHandler *NotifyHandler) *ginext.Engine {
	router := ginext.New("release")

	router.POST("/notify", notifyHandler.CreateNotification)
	router.GET("/notify/:id", notifyHandler.GetNotification)
	router.DELETE("/notify/:id", notifyHandler.DeleteNotification)
	return router
}
