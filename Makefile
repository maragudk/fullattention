.PHONY: benchmark
benchmark:
	go test -bench . ./...

.PHONY: build-css
build-css: tailwindcss
	./tailwindcss -i tailwind.css -o public/styles/app.css --minify

.PHONY: build-docker
build-docker:
	docker build --platform linux/arm64,linux/amd64 -t full-attention .

.PHONY: clean-all
clean-all:
	rm -f app.db*

.PHONY: cover
cover:
	go tool cover -html cover.out

.PHONY: fmt
fmt:
	goimports -w -local `head -n 1 go.mod | sed 's/^module //'` .

.PHONY: lint
lint:
	golangci-lint run

tailwindcss:
	curl -sL -o tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
	chmod a+x tailwindcss

.PHONY: test
test:
	go test -coverprofile cover.out -shuffle on ./...

.PHONY: watch
watch:
	./watch.sh

.PHONY: watch-css
watch-css: tailwindcss
	./tailwindcss -i tailwind.css -o public/styles/app.css --watch
