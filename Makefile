.PHONY: dev build tidy test infra-init infra-plan infra-apply infra-destroy deploy

dev:
	docker compose up --build

tidy:
	go mod tidy

build:
	docker build --platform linux/amd64 -t microservice:latest .

test:
	go test ./...

infra-init:
	terraform -chdir=infra init

infra-plan:
	terraform -chdir=infra plan

infra-apply:
	terraform -chdir=infra apply

infra-destroy:
	terraform -chdir=infra destroy

deploy:
	chmod +x deploy.sh && ./deploy.sh
