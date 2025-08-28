package html

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type HomePageProps struct {
	PageProps
}

func HomePage(props HomePageProps) Node {
	return Page(props.PageProps,
		H1(Text("Full Attention")),
	)
}
