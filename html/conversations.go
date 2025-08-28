package html

import (
	"strings"

	"github.com/yuin/goldmark"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"

	"app/model"
)

func ConversationsPage(props PageProps, cd model.ConversationDocument) Node {
	props.Title = cd.Conversation.Topic
	if props.Title == "" {
		props.Title = cd.Conversation.ID.String()
	}

	return Page(props,
		Group{
			H1(Text(props.Title)),

			Div(Class("space-y-8"), hx.Get("/conversations?id="+cd.Conversation.ID.String()), hx.Trigger("every 1s"),
				TurnsPartial(cd),
			),
		},
	)
}

func TurnsPartial(cd model.ConversationDocument) Node {
	return Map(cd.Turns, func(t model.Turn) Node {
		s := cd.Speakers[t.SpeakerID]

		var content string
		var b strings.Builder
		if err := goldmark.Convert([]byte(t.Content), &b); err != nil {
			content = "Error converting markdown to HTML: " + err.Error()
		} else {
			content = b.String()
		}

		return Div(Class("flex"),
			P(Text(s.Name)),
			Div(Class("border border-gray-200 rounded-lg w-full px-4 mx-4"), Raw(content)),
		)
	})
}
