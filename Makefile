all: build

build:
	bash xcaddy.sh

.PHONY: frontend
frontend:
	cd frontend && rm -rf dist && npm run build
	rm -rf static/assets && cp -R frontend/dist/. static
