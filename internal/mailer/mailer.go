package mailer

import "embed"

const (
	FromName    = "GoBlog"
	maxRetries  = 3
	UserWelcome = "user_invitation.templ"
)

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any, isSandBox bool) error
}
