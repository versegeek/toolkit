.ONESHELL:

SHELL := /bin/sh

GOPROXY ?= https://goproxy.cn,direct

# Wire runs at a pinned version without adding the generator to go.mod.
WIRE_VERSION = v0.7.0
WIRE = go run github.com/google/wire/cmd/wire@$(WIRE_VERSION)
# Packages that contain Wire injectors.
WIRE_PKGS = ./cmd/
