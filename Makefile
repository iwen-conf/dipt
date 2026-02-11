APP      := dipt
CMD      := .
DIST     := dist
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X main.version=$(VERSION)

# 目标平台: OS/ARCH
PLATFORMS := \
	linux/amd64 linux/arm64 linux/arm linux/386 \
	darwin/amd64 darwin/arm64 \
	windows/amd64 windows/arm64

.PHONY: all build clean release $(PLATFORMS)

# 默认: 编译当前平台
build:
	@mkdir -p $(DIST)
	go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP) $(CMD)

# 交叉编译所有平台
all: clean $(PLATFORMS)

$(PLATFORMS):
	$(eval OS   := $(word 1,$(subst /, ,$@)))
	$(eval ARCH := $(word 2,$(subst /, ,$@)))
	$(eval EXT  := $(if $(filter windows,$(OS)),.exe,))
	@echo "=> $(OS)/$(ARCH)"
	@mkdir -p $(DIST)
	GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags "$(LDFLAGS)" \
		-o $(DIST)/$(APP)-$(OS)-$(ARCH)$(EXT) $(CMD)

# 打包为 tar.gz / zip
release: all
	@cd $(DIST) && for f in $(APP)-linux-* $(APP)-darwin-*; do \
		[ -f "$$f" ] && tar czf "$$f.tar.gz" "$$f" && rm "$$f"; \
	done; \
	for f in $(APP)-windows-*; do \
		[ -f "$$f" ] && zip -q "$$f.zip" "$$f" && rm "$$f"; \
	done
	@echo "打包完成 → $(DIST)/"

clean:
	rm -rf $(DIST) $(APP)

test:
	go test ./...
