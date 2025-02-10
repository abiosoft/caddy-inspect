all: build

build:
	bash xcaddy.sh

.PHONY: frontend
frontend:
	cd frontend && rm -rf dist && npm run build && cp -R dist/. ../static
