package dto

import (
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/types"
	"github.com/wb-go/wbf/ginext"
)


type GetNotificationRequest struct {
	ID string `uri:"id" binding:"required,uuid"`
}

func (r *GetNotificationRequest) ToUUID() (types.UUID, error) {
	return types.NewUUID(r.ID)
}

func BindNotificationRequest(g * ginext.Context) (*GetNotificationRequest, error) {
	var req *GetNotificationRequest
	err := g.BindUri(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}