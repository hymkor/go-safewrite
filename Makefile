ifeq ($(OS),Windows_NT)
    SHELL=CMD.EXE
    SET=set
    WHICH=where.exe
    DEL=del
    NUL=nul
else
    SET=export
    WHICH=which
    DEL=rm
    NUL=/dev/null
endif

ifndef GO
    SUPPORTGO=go1.20.14
    GO:=$(shell $(WHICH) $(SUPPORTGO) 2>$(NUL) || echo go)
endif

NAME:=$(notdir $(CURDIR))
VERSION:=$(shell git describe --tags 2>$(NUL) || echo v0.0.0)
GOOPT:=-ldflags "-s -w -X main.version=$(VERSION)"
EXE:=$(shell $(GO) env GOEXE)

all:
	$(GO) fmt ./...
	$(SET) "CGO_ENABLED=0" && $(GO) build $(GOOPT)

readme:
	$(GO) run github.com/hymkor/example-into-readme@latest
	$(GO) run github.com/hymkor/example-into-readme@latest -target README_ja.md

demo:
	$(GO) run example.go

.PHONY: all readme demo
