.PHONY: build test tidy fmt

GO := GOWORK=$(CURDIR)/go.work GOEXPERIMENT=jsonv2 go

MODULES := \
	github.com/khiemnguyen15/twitch-watcher/pkg/... \
	github.com/khiemnguyen15/twitch-watcher/services/subscription-service/... \
	github.com/khiemnguyen15/twitch-watcher/services/stream-poller/... \
	github.com/khiemnguyen15/twitch-watcher/services/stream-filter/... \
	github.com/khiemnguyen15/twitch-watcher/services/notification-dispatcher/...

build:
	$(GO) build $(MODULES)

test:
	$(GO) test $(MODULES)

tidy:
	$(GO) work sync
	@for mod in pkg services/subscription-service services/stream-poller services/stream-filter services/notification-dispatcher; do \
		echo "→ tidy $$mod"; \
		(cd $$mod && GOWORK=$(CURDIR)/go.work GOEXPERIMENT=jsonv2 go mod tidy); \
	done

fmt:
	@for mod in pkg services/subscription-service services/stream-poller services/stream-filter services/notification-dispatcher; do \
		(cd $$mod && GOWORK=$(CURDIR)/go.work GOEXPERIMENT=jsonv2 gofmt -w .); \
	done

.PHONY: run-subscription-service
run-subscription-service:
	$(GO) run github.com/khiemnguyen15/twitch-watcher/services/subscription-service/cmd

.PHONY: run-stream-poller
run-stream-poller:
	$(GO) run github.com/khiemnguyen15/twitch-watcher/services/stream-poller/cmd

.PHONY: run-stream-filter
run-stream-filter:
	$(GO) run github.com/khiemnguyen15/twitch-watcher/services/stream-filter/cmd

.PHONY: run-notification-dispatcher
run-notification-dispatcher:
	$(GO) run github.com/khiemnguyen15/twitch-watcher/services/notification-dispatcher/cmd
