package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/dto"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/ports"
	"github.com/wb-go/wbf/ginext"
)

type NotifyHandler struct {
	crudService ports.CRUDServiceInterface
}

func NewNotifyHandler(crudService ports.CRUDServiceInterface) *NotifyHandler {
	return &NotifyHandler{crudService: crudService}
}

func (h *NotifyHandler) CreateNotification(c *ginext.Context) {
	var body dto.NotificationCreate

	err := c.BindJSON(&body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": fmt.Sprintf("invalid body (parsing): %s", err.Error())})
		return
	}

	var createModel *model.Notification
	createModel, err = body.ToEnity()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ginext.H{"error": fmt.Sprintf("invalid body (validating): %s", err.Error())})
		return
	}

	_, err = h.crudService.CreateNotification(context.Background(), createModel)
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusConflict,
			ginext.H{"error": fmt.Sprintf("couldn't perform operation: %s", err.Error())},
		)
		return
	}
	c.JSON(http.StatusCreated, dto.ToFullFromModelNotification(createModel))
}

func (h *NotifyHandler) GetNotification(c *ginext.Context) {
	req, err := dto.BindNotificationRequest(c)
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			ginext.H{"error": fmt.Sprintf("invalid ID parameter: %s", err.Error())},
		)
		return
	}

	id, err := req.ToUUID()
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			ginext.H{"error": fmt.Sprintf("invalid UUID format: %s", err.Error())},
		)
		return
	}
	notification, err := h.crudService.GetNotification(context.Background(), id)
	if err != nil {
		// if errors.Is(err, internalerrors.ErrNotificationNotFound) {
		// 	c.AbortWithStatusJSON(
		// 		http.StatusNotFound,
		// 		gin.H{"error": "notification not found"},
		// 	)
		// 	return
		// }

		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			ginext.H{"error": fmt.Sprintf("couldn't get notification: %s", err.Error())},
		)
		return
	}

	c.JSON(http.StatusOK, notification)
}

func (h *NotifyHandler) DeleteNotification(c *ginext.Context) {
	req, err := dto.BindNotificationRequest(c)
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			ginext.H{"error": fmt.Sprintf("invalid ID parameter: %s", err.Error())},
		)
		return
	}

	id, err := req.ToUUID()
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			ginext.H{"error": fmt.Sprintf("invalid UUID format: %s", err.Error())},
		)
		return
	}

	err = h.crudService.DeleteNotification(context.Background(), id)
	if err != nil {
		// if errors.Is(err, internalerrors.ErrNotificationNotFound) {
		// 	c.AbortWithStatusJSON(
		// 		http.StatusNotFound,
		// 		gin.H{"error": "notification not found"},
		// 	)
		// 	return
		// }

		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			ginext.H{"error": fmt.Sprintf("couldn't delete notification: %s", err.Error())},
		)
		return
	}

	c.Status(http.StatusNoContent)
}
