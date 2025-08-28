package http

import (
	"context"
	"log/slog"

	. "maragu.dev/gomponents"

	"app/html"
	"app/model"
)

type conversationsGetter interface {
	GetConversations(ctx context.Context) ([]model.Conversation, error)
}

func Home(r *Router, log *slog.Logger, db conversationsGetter) {
	r.Get("/", func(props html.PageProps) (Node, error) {
		cs, err := db.GetConversations(props.Ctx)
		if err != nil {
			log.Info("Error getting conversations", "error", err)
			return html.ErrorPage(), err
		}

		return html.HomePage(html.HomePageProps{PageProps: props}, cs), nil
	})
}
