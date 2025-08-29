package html

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"maragu.dev/glue/html"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

var hashOnce sync.Once
var appCSSPath, htmxJSPath, idiomorphJSPath, idiomorphExtJSPath, appJSPath string

func Page(props PageProps, body ...Node) Node {
	hashOnce.Do(func() {
		appCSSPath = getHashedPath("public/styles/app.css")
		htmxJSPath = getHashedPath("public/scripts/htmx.js")
		idiomorphJSPath = getHashedPath("public/scripts/idiomorph.js")
		idiomorphExtJSPath = getHashedPath("public/scripts/idiomorph-ext.js")
		appJSPath = getHashedPath("public/scripts/app.js")
	})

	title := props.Title
	if title != "" {
		title += " - "
	}
	title += "Full Attention"

	return HTML5(HTML5Props{
		Title:       title,
		Description: props.Description,
		Language:    "en",
		Head: []Node{
			Link(Rel("stylesheet"), Href(appCSSPath)),
			Script(Src(htmxJSPath), Defer()),
			Script(Src(idiomorphJSPath), Defer()),
			Script(Src(idiomorphExtJSPath), Defer()),
			Script(Src(appJSPath), Defer()),
			Meta(Name("htmx-config"), Content(`{"scrollIntoViewOnBoost":false}`)),
			html.FavIcons("Full Attention"),
		},
		Body: []Node{Class("bg-primary-800 text-warm-gray-900 dark:text-white font-mono"),
			hx.Ext("morph"),
			Div(Class("min-h-dvh flex flex-col justify-between"),
				header(props),
				Div(Class("grow bg-white dark:bg-warm-gray-800 h-auto"),
					container(true,
						Group(body),
					),
				),
			),
		},
	})
}

func header(_ PageProps) Node {
	return Div(
		container(false),
	)
}

func container(padY bool, children ...Node) Node {
	return Div(
		Classes{
			"max-w-7xl mx-auto h-full": true,
			"px-4 sm:px-6 lg:px-8":     true,
			"py-4 md:py-8":             padY,
		},
		Group(children),
	)
}

func getHashedPath(path string) string {
	externalPath := strings.TrimPrefix(path, "public")
	ext := filepath.Ext(path)
	if ext == "" {
		panic("no extension found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("%v.x%v", strings.TrimSuffix(externalPath, ext), ext)
	}

	return fmt.Sprintf("%v.%x%v", strings.TrimSuffix(externalPath, ext), sha256.Sum256(data), ext)
}
