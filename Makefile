.PHONY: tailwind-watch
tailwind-watch:
	./tailwindcss -i ./static/css/input.css -o ./static/css/style.css --watch

.PHONY: tailwind-build
tailwind-build:
	./tailwindcss -i ./static/css/input.css -o ./static/css/style.min.css --minify

.PHONY: templ-generate
templ-generate:
	templ generate

.PHONY: templ-watch
templ-watch:
	templ generate --watch
	
.PHONY: dev
dev:
	go build -o ./bin/app ./cmd && go build -o ./bin/agent ./cmd/agent
	PORT=:4101 AGENT_NAME=lab-pc-1 AGENT_TOKEN=dev-token AGENT_SERVER=http://localhost:4000 AGENT_URL=http://localhost:4101 ./bin/agent & AGENT1=$$!; \
	PORT=:4102 AGENT_NAME=lab-pc-2 AGENT_VERSION=2.4.1 AGENT_TOKEN=dev-token AGENT_SERVER=http://localhost:4000 AGENT_URL=http://localhost:4102 ./bin/agent & AGENT2=$$!; \
	trap "kill $$AGENT1 $$AGENT2 2>/dev/null" EXIT INT TERM; \
	ADMIN_EMAIL=admin@localhost AGENT_TOKEN=dev-token SESSION_SECRET=dev-secret ./bin/app

.PHONY: build
build:
	make templ-generate
	go build -o ./bin/app ./cmd
	go build -o ./bin/agent ./cmd/agent

.PHONY: vet
vet:
	go vet ./...

.PHONY: staticcheck
staticcheck:
	staticcheck ./...

.PHONY: test
test:
	  go test -race -v -timeout 30s ./...