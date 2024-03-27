# SPDX-FileCopyrightText: 2022 Risk.Ident GmbH <contact@riskident.com>
#
# SPDX-License-Identifier: CC0-1.0

.PHONY: all
all: jelease.schema.json generate

jelease.schema.json: pkg/config/*.go cmd/config_schema.go
	go run . config schema --output jelease.schema.json

.PHONY: generate
generate:
	go run github.com/a-h/templ/cmd/templ@$(shell go list -m -f '{{ .Version }}' github.com/a-h/templ) generate

.PHONY: test
test:
	go test ./...

.PHONY: deps
deps: deps-npm deps-pip

.PHONY: deps-pip
deps-pip:
	pip install --user reuse

.PHONY: deps-npm
deps-npm: node_modules

node_modules: package.json
	npm install

.PHONY: lint
lint: lint-md lint-license

.PHONY: lint-fix
lint-fix: lint-md-fix

.PHONY: lint-md
lint-md: node_modules
	npx remark .

.PHONY: lint-md-fix
lint-md-fix: node_modules
	npx remark . -o

.PHONY: lint-license
lint-license:
	reuse lint
