TEMPL_VERSION    := v0.3.1020
TAILWIND_VERSION := v4.3.3
TAILWIND         := bin/tailwindcss
TEMPL            := go run github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION)

.PHONY: generate css build run test vet fmt tailwind

generate: ## templ -> templates/*_templ.go
	$(TEMPL) generate

css: $(TAILWIND) ## static/css/input.css -> static/css/app.css
	$(TAILWIND) -i static/css/input.css -o static/css/app.css --minify

build: generate css
	go build -o app .

run: generate css
	go run .

test:
	go test ./... -race -cover

vet:
	go vet ./...

fmt:
	$(TEMPL) fmt templates
	go fmt ./...

# Downloads the standalone Tailwind CLI for this machine (no Node needed).
$(TAILWIND):
	@mkdir -p bin
	@os=$$(uname -s | tr A-Z a-z | sed 's/darwin/macos/'); arch=$$(uname -m | sed 's/x86_64/x64/;s/aarch64/arm64/'); \
	curl -sSfL -o $(TAILWIND) https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-$$os-$$arch
	@chmod +x $(TAILWIND)
