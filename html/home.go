package html

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"

	"app/model"
)

type HomePageProps struct {
	PageProps
}

func HomePage(props HomePageProps, cs []model.Conversation) Node {
	return Page(props.PageProps,
		Ol(
			Map(cs, func(c model.Conversation) Node {
				linkText := c.Topic
				if linkText == "" {
					linkText = c.ID.String()
				}
				return Li(A(Href("/conversations?id="+c.ID.String()), Text(linkText)))
			}),
		),
	)
}
