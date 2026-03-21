GO := ~/.local/go/current/bin/go

.PHONY: build test run-iam run-platform run-education

build:
	cd /Users/chenrui/Desktop/go-migration-platform && $(GO) build ./...

test:
	cd /Users/chenrui/Desktop/go-migration-platform && $(GO) test ./...

run-iam:
	cd /Users/chenrui/Desktop/go-migration-platform && $(GO) run ./services/iam/cmd/api

run-platform:
	cd /Users/chenrui/Desktop/go-migration-platform && $(GO) run ./services/platform/cmd/api

run-education:
	cd /Users/chenrui/Desktop/go-migration-platform && $(GO) run ./services/education/cmd/api
