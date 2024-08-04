# Detect the operating system
OS := $(shell uname -s)

# Define command variables based on the OS
ifeq ($(OS), Linux)
    TAILWIND := ./tailwindcss
    GO_BUILD := CGO_ENABLED=1 go build -o ./tmp/$(APP_NAME) ./cmd/$(APP_NAME)/main.go
    GO_BUILD_PROD := CGO_ENABLED=1 go build -ldflags "-X main.Environment=production" -o ./bin/$(APP_NAME) ./cmd/$(APP_NAME)/main.go
else ifeq ($(OS), Darwin)
    TAILWIND := ./tailwindcss
    GO_BUILD := CGO_ENABLED=1 go build -o ./tmp/$(APP_NAME) ./cmd/$(APP_NAME)/main.go
    GO_BUILD_PROD := CGO_ENABLED=1 go build -ldflags "-X main.Environment=production" -o ./bin/$(APP_NAME) ./cmd/$(APP_NAME)/main.go
else
    TAILWIND := tailwindcss
    GO_BUILD := set CGO_ENABLED=1 && go build -o .\tmp\$(APP_NAME).exe .\cmd\$(APP_NAME)\main.go
    GO_BUILD_PROD := set CGO_ENABLED=1 && go build -ldflags "-X main.Environment=production" -o .\bin\$(APP_NAME).exe .\cmd\$(APP_NAME)\main.go
endif

.PHONY: tailwind-watch
tailwind-watch:
	$(TAILWIND) -i ./static/css/input.css -o ./static/css/style.css --watch

.PHONY: tailwind-build
tailwind-build:
	$(TAILWIND) -i ./static/css/input.css -o ./static/css/style.min.css --minify

.PHONY: templ-generate
templ-generate:
	templ generate

.PHONY: templ-watch
templ-watch:
	templ generate --watch
	
.PHONY: dev
dev:
	$(GO_BUILD) && air

.PHONY: build
build:
	make tailwind-build
	make templ-generate
	$(GO_BUILD_PROD)

.PHONY: vet
vet:
	go vet ./...

.PHONY: staticcheck
staticcheck:
	staticcheck ./...

.PHONY: test
test:
	go test -race -v -timeout 30s ./...
