package mailer

import (
	"context"
	"embed"
)

const (
	FromName            = "GopherSocial"
	maxRetires          = 3
	UserWelcomeTemplate = "user_invitation.tmpl"
)

//go:embed "templates"
var FS embed.FS

type MailClient interface {
	Send(ctx context.Context, templateFile, email string, data any) (int, error)
}
