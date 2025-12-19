package internaltypes

import (
	"fmt"
	"net/mail"
	"strconv"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/types"
)

var ErrInvalidNotificationChannelValue = fmt.Errorf("invalid notification channel value: possible ones are: '%s', '%s', '%s'")

const (
	// EMAIL is the constant value for email channel string value
	EMAIL = "email"
	// TELEGRAM is the constant value for telegram channel string value
	TELEGRAM = "telegram"
	// CONSOLE is the constant value for console channel string value
	CONSOLE = "console"
)

var (
	// ChannelEmail is an example channel with value EMAIL
	ChannelEmail    = NotificationChannel{val: EMAIL}
	ChannelTelegram = NotificationChannel{val: TELEGRAM}
	ChannelConsole  = NotificationChannel{val: CONSOLE}
)

type NotificationChannel struct {
	val types.AnyText
}

func(c NotificationChannel) String() string {
	return c.val.String()
}

func NotificationChannelFromString(Val string) (NotificationChannel, error) {
	switch Val {
	case EMAIL, TELEGRAM, CONSOLE:
		break
	default:
		return NotificationChannel{}, ErrInvalidNotificationChannelValue
	}
	return NotificationChannel{val: types.NewAnyText(Val)}, nil
}

type Recipient struct {
	Val types.AnyText
}
func(c Recipient) String() string {
	return c.Val.String()
}

func RecipientFromString(val string) Recipient {
	return Recipient{
		Val: types.NewAnyText(val),
	}
}

func NewSendTo(Val types.AnyText, channel NotificationChannel) (Recipient, error) {
	switch channel {
	case ChannelEmail:
		_, err := mail.ParseAddress(Val.String())
		if err != nil {
			return Recipient{}, fmt.Errorf("invalid email address: %w", err)
		}
	case ChannelTelegram:
		_, err := strconv.ParseInt(Val.String(), 10, 64)
		if err != nil {
			return Recipient{}, fmt.Errorf("invalid telegram address: %s", Val.String())
		}
	default:
		break
	}

	return Recipient{Val: Val}, nil
}
