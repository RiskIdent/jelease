# SPDX-FileCopyrightText: 2022 Risk.Ident GmbH <contact@riskident.com>
#
# SPDX-License-Identifier: CC0-1.0

.PHONY: deps
deps: deps-npm deps-pip

.PHONY: deps-pip
deps-pip:
	pip install --user yamllint ansible-lint reuse

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
