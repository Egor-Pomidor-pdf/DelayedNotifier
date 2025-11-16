package notificationheap

import (
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/model"
)

type NotificationHeap []*model.Notification

func (h *NotificationHeap) Len() int { return len(*h) }

func (h *NotificationHeap) Less(i, j int) bool {
	timeI, _ := time.Parse(time.RFC3339, (*h)[i].ScheduledAt.String())
	timeJ, _ := time.Parse(time.RFC3339, (*h)[j].ScheduledAt.String())
	return timeI.After(timeJ)
}

func (h *NotificationHeap) Swap(i, j int) { (*h)[i], (*h)[j] = (*h)[j], (*h)[i] }

func (h *NotificationHeap) Push(x interface{}) {
	*h = append(*h, x.(*model.Notification))
}

func (h *NotificationHeap) Pop() any {
	old := *h
	n := len(*h)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func (h *NotificationHeap) Peek() *model.Notification {
	if h.Len() == 0 {
		return nil
	}
	n := len(*h)
	return (*h)[n-1]
}