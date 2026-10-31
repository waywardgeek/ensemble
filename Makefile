# Edition 3 C# application; Go is retained for grading and publishing tools.
.DEFAULT_GOAL := all
DOTNET ?= dotnet
DOTNET_PATH := $(shell command -v $(DOTNET) 2>/dev/null)
GO ?= go
CH ?= 1

.PHONY: all build ensemble cli clean test test-csharp test-go test-legacy-go vet fmt lint-prose grade grade-dir fake fake-serve book ebook print
all: build
build: ensemble

ensemble:
	$(DOTNET) build agent/Examples/Gui/Gui.csproj
	@printf '%s\n' '#!/bin/sh' 'set -eu' 'export DOTNET_HOST_PATH="$${DOTNET_HOST_PATH:-$(DOTNET_PATH)}"' 'here=$$(CDPATH= cd -- "$$(dirname -- "$$0")" && pwd)' 'exec "$$here/agent/start" "$$@"' > ensemble
	@chmod +x ensemble

cli:
	$(DOTNET) build agent/Cli/Ch04.csproj

clean:
	$(DOTNET) clean agent/Examples/Gui/Gui.csproj
	$(DOTNET) clean agent/Cli/Ch04.csproj
	rm -f ensemble

test: test-csharp test-go
test-csharp:
	$(DOTNET) run --project agent/Tests/Tests.csproj --no-launch-profile

test-go: vet
	$(GO) test ./cmd/... ./internal/gradeedition3 ./internal/fakevendor ./internal/fakeanthropic

# Historical source-layout/pricing expectations are not V3 acceptance.
test-legacy-go:
	$(GO) test ./cmd/... ./internal/...
vet:
	$(GO) vet ./cmd/... ./internal/...
fmt:
	$(GO) fmt ./cmd/... ./internal/...
lint-prose:
	$(GO) run ./cmd/lintprose $(LINTFLAGS) book/preface.md book/chapter-[0-9][0-9].md

# Always supply an explicit C# executable/wrapper and chapter-specific flags.
grade grade-dir:
	@test -n "$(DIR)" || (echo 'Set DIR to the explicit submission or CLI wrapper; see the chapter grading scope' >&2; exit 2)
	$(GO) run ./cmd/grade-edition3 -ch $(CH) $(GRADEFLAGS) $(DIR)

VENDOR ?= anthropic
FAKE_CH ?= 3
fake:
	$(GO) run ./cmd/fakevendor -ch $(FAKE_CH) -vendor $(VENDOR) chat
fake-serve:
	$(GO) run ./cmd/fakevendor -vendor $(VENDOR)

book:
	scripts/assemble-book.sh > book/the-art-of-building-ai-coding-agents.md
ebook: book
	scripts/build-ebook.sh
print: book
	python3 book/build-print.py
