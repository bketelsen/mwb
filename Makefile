.PHONY: build install uninstall clean fmt format-check tidy-check vet lint-version-check lint test race cross verify check ci vulncheck goreleaser-version goreleaser-version-check release-check svu-version-check bump-main-check bump site site-serve

SYSTEMD_USER_DIR := $(HOME)/.config/systemd/user
GO_FILES = $(shell git ls-files --cached --others --exclude-standard '*.go')
GOLANGCI_LINT_VERSION := $(strip $(shell sed -n 's/^golangci-lint = "\(.*\)"/\1/p' mise.toml))
SVU_VERSION := $(strip $(shell sed -n 's/^svu = "\(.*\)"/\1/p' mise.toml))
GORELEASER_VERSION := $(strip $(shell sed -n 's/^"github:goreleaser\/goreleaser-pro" = "\(.*\)"/\1/p' mise.toml))
GO_VERSION := $(strip $(shell sed -n 's/^go \([0-9.]*\)$$/\1/p' go.mod))

build:
	go build -o mwb ./cmd/mwb

install: build
	install -D mwb $(HOME)/go/bin/mwb
	install -d $(SYSTEMD_USER_DIR)
	install -m 644 mwb.service $(SYSTEMD_USER_DIR)/mwb.service
	systemctl --user daemon-reload
	@echo ""
	@echo "Installed. To start:"
	@echo "  systemctl --user enable --now mwb"
	@echo ""
	@echo "View logs:"
	@echo "  journalctl --user -u mwb -f"

uninstall:
	systemctl --user disable --now mwb || true
	rm -f $(SYSTEMD_USER_DIR)/mwb.service
	rm -f $(HOME)/go/bin/mwb
	systemctl --user daemon-reload

clean:
	rm -f mwb
	rm -rf build/

fmt:
	gofmt -w $(GO_FILES)

format-check:
	@files="$$(gofmt -l $(GO_FILES))"; \
	if [ -n "$$files" ]; then \
		printf 'Files need gofmt:\n%s\n' "$$files"; exit 1; \
	fi

tidy-check:
	go mod tidy -diff

vet:
	go vet ./...

lint-version-check:
	@test -n "$(GOLANGCI_LINT_VERSION)" || { echo "mise.toml pins no golangci-lint (install with: mise install)"; exit 1; }
	@installed="$$(golangci-lint version --short 2>/dev/null)" || { \
		echo "golangci-lint $(GOLANGCI_LINT_VERSION) is required (install with: mise install)"; exit 1; }; \
	if [ "$$installed" != "$(GOLANGCI_LINT_VERSION)" ]; then \
		echo "expected golangci-lint $(GOLANGCI_LINT_VERSION), found $$installed (install with: mise install)"; exit 1; \
	fi; \
	built="$$(golangci-lint version 2>/dev/null | sed -n 's/.*built with go\([0-9.]*\).*/\1/p')"; \
	built_minor="$$(printf '%s\n' "$$built" | sed -n 's/^\([0-9][0-9]*\.[0-9][0-9]*\).*/\1/p')"; \
	go_minor="$$(printf '%s\n' "$(GO_VERSION)" | sed -n 's/^\([0-9][0-9]*\.[0-9][0-9]*\).*/\1/p')"; \
	if [ -n "$$built_minor" ] && [ -n "$$go_minor" ] && [ "$$(printf '%s\n%s\n' "$$go_minor" "$$built_minor" | sort -V | head -1)" != "$$go_minor" ]; then \
		echo "golangci-lint $(GOLANGCI_LINT_VERSION) was built with go$$built, older than go.mod's go$(GO_VERSION) (install with: mise install)"; exit 1; \
	fi

lint: lint-version-check
	golangci-lint run

test:
	go test ./...

race:
	go test -race ./...

cross:
	mkdir -p build
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o build/mwb-linux-amd64 ./cmd/mwb
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -o build/mwb-linux-arm64 ./cmd/mwb

# The gate triad of frostyard/core ADR-0043 and ADR-0044:
# verify changes no file, check can format, and ci adds the CI-only checks.
verify: tidy-check vet format-check lint test

check: fmt verify

# ci.yml runs `make ci`; vulncheck stays outside because it needs the network
# (frostyard/core ADR-0083).
ci: verify race build cross release-check

vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...

goreleaser-version:
	@test -n "$(GORELEASER_VERSION)" || { echo "mise.toml pins no goreleaser-pro (install with: mise install)" >&2; exit 1; }
	@printf 'v%s\n' "$(GORELEASER_VERSION)"

goreleaser-version-check:
	@test -n "$(GORELEASER_VERSION)" || { echo "mise.toml pins no goreleaser-pro (install with: mise install)"; exit 1; }
	@installed="$$(goreleaser --version 2>/dev/null | sed -n 's/^GitVersion:[[:space:]]*//p')"; \
	if [ "$$installed" != "$(GORELEASER_VERSION)" ]; then \
		echo "expected goreleaser-pro $(GORELEASER_VERSION), found $${installed:-missing} (install with: mise install)"; exit 1; \
	fi

# --soft warns on deprecations (issue #20) but still fails on invalid config.
release-check: goreleaser-version-check
	goreleaser check --soft
	@if output="$$(goreleaser check --soft testdata/goreleaser-invalid.yml 2>&1)"; then \
		printf '%s\n' "$$output"; \
		echo "goreleaser check accepted testdata/goreleaser-invalid.yml"; exit 1; \
	elif ! printf '%s\n' "$$output" | grep -Fq 'bogus_field'; then \
		printf '%s\n' "$$output"; \
		echo "goreleaser check did not reject bogus_field in testdata/goreleaser-invalid.yml"; exit 1; \
	fi

svu-version-check:
	@test -n "$(SVU_VERSION)" || { echo "mise.toml pins no svu (install with: mise install)"; exit 1; }
	@installed="$$(svu --version 2>/dev/null | sed -n 's/^GitVersion:[[:space:]]*//p')"; \
	if [ "$$installed" != "$(SVU_VERSION)" ]; then \
		echo "expected svu $(SVU_VERSION), found $$installed (install with: mise install)"; exit 1; \
	fi

bump-main-check:
	@test -z "$$(git status --porcelain)" || { echo "The working tree is not clean; commit or stash before bumping."; exit 1; }
	@test "$$(git branch --show-current)" = main || { echo "Bump from main."; exit 1; }
	git fetch --quiet --tags origin main
	@test "$$(git rev-parse HEAD)" = "$$(git rev-parse origin/main)" || { echo "HEAD is not origin/main; pull or push first."; exit 1; }

bump:
	$(MAKE) svu-version-check
	$(MAKE) bump-main-check
	$(MAKE) ci
	@test -z "$$(git status --porcelain)" || { echo "make ci changed the working tree; not tagging."; exit 1; }
	@version=$$(svu next); \
		git tag -a "$$version" -m "Version $$version"; \
		echo "Tagged $$version; pushing it to origin"; \
		git push origin "$$version"

# The documentation site (site/), built with pinned ProperDocs and MaterialX.
SITE_VENV := build/site-venv

$(SITE_VENV)/bin/properdocs: site/requirements.txt
	python3 -m venv $(SITE_VENV)
	$(SITE_VENV)/bin/pip install -q --require-hashes -r site/requirements.txt
	touch $@

site: $(SITE_VENV)/bin/properdocs
	$(SITE_VENV)/bin/properdocs build --strict -f site/properdocs.yml

site-serve: $(SITE_VENV)/bin/properdocs
	$(SITE_VENV)/bin/properdocs serve -f site/properdocs.yml

