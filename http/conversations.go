package http

import (
	"context"
	"log/slog"
	"net/http"

	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx/http"

	"app/html"
	"app/model"
)

type conversationGetter interface {
	GetConversationDocument(ctx context.Context, id model.ConversationID) (model.ConversationDocument, error)
}

func Conversations(r *Router, log *slog.Logger, db conversationGetter) {
	r.Get("/conversations", func(props html.PageProps) (Node, error) {
		id := model.ConversationID(props.R.URL.Query().Get("id"))

		if id == "" {
			http.Error(props.W, "id is required", http.StatusBadRequest)
			return nil, nil
		}

		cd, err := db.GetConversationDocument(props.Ctx, id)
		if err != nil {
			log.Info("Error getting conversation document", "error", err)
			return html.ErrorPage(), err
		}

		if hx.IsRequest(props.R.Header) {
			return html.TurnsPartial(cd), nil
		}

		return html.ConversationsPage(props, cd), nil
	})
}
