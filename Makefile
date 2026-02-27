TOOLS   := pin ask diff stand
DIST    := dist
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build clean tidy lint $(addprefix build-,$(TOOLS))

build: $(addprefix build-,$(TOOLS))

$(addprefix build-,$(TOOLS)): build-%:
	@mkdir -p $(DIST)
	go build \
		-ldflags="-X github.com/reky0/glyph-$*/cmd.Version=$(VERSION)" \
		-o $(DIST)/$* \
		./tools/$*/cmd/$*
	@echo "built $(DIST)/$*"

clean:
	rm -rf $(DIST)

tidy:
	@for mod in libs/glyph-core libs/glyph-ink libs/glyph-store libs/glyph-mind \
	            tools/pin tools/ask tools/diff tools/stand; do \
	  echo "tidy $$mod"; \
	  (cd $$mod && go mod tidy); \
	done

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
	  golangci-lint run ./...; \
	else \
	  echo "golangci-lint not found, skipping"; \
	fi
