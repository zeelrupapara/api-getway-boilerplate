GOPATH:=$(shell go env GOPATH)

.PHONY: init
init:
	@go get -u google.golang.org/protobuf@v1.26.0 
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest	
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	
.PHONY: proto
proto:
	 protoc --go_out=. --go_opt=paths=source_relative \
     --go-grpc_out=. --go-grpc_opt=paths=source_relative \
     proto/greenlync/*.proto

.PHONY: update
update:
	@go get -u

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: test
test:
	@go test -v ./... -cover

.PHONY: swagger
swagger:
	@swag fmt -g cmd/main.go
	@swag init -g cmd/main.go

.PHONY: docker-stack-up
docker-stack-up:
	@docker compose -f stack.yaml up -d

.PHONY: docker-stack-down
docker-stack-down:
	@docker compose -f stack.yaml down

.PHONY: docker-compose-up
docker-compose-up:
	@docker rmi greenlync-api-gateway:dev
	@docker compose up -d

.PHONY: docker-compose-down
docker-compose-down:
	@docker compose down

.PHONY: run
run:
	@go run cmd/main.go

.PHONY: seed
seed:
	@go run data/*.go -user greenlync -password

.PHONY: kill
kill:
	kill -9 $(lsof -t -i tcp:3000)

.PHONY: race
race:
	@CGO_ENABLED=1 go run -race cmd/main.go	

.PHONY: web
web:
	@sudo rm -rf public/web
	@mkdir public/web
	@echo "Web assets build - configure your frontend build process" 	

.PHONY: docs
docs:
	@echo "Documentation build - configure your docs build process" 	

.PHONY: build-app
build-app:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o ./cmd/greenlync-api-gateway cmd/*.go

.PHONY: build-docker-image
build-docker-image:
	@docker rmi greenlync-api-gateway:dev
	@docker build -t greenlync-api-gateway:dev .	

.PHONY: push-dev
push-dev:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o ./cmd/greenlync-api-gateway cmd/*.go
	@sudo docker build -t greenlync-api-gateway:dev .
	@echo "Configure your container registry for push" 

.PHONY: push-staging
push-staging:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o ./cmd/greenlync-api-gateway cmd/*.go
	@sudo docker build -t greenlync-api-gateway:staging .
	@echo "Configure your container registry for push"

.PHONY: push-prod
push-prod:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o ./cmd/greenlync-api-gateway cmd/*.go
	@sudo docker build -t greenlync-api-gateway:prod .
	@echo "Configure your container registry for push"	

.PHONY: deploy-dev
deploy-dev:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o ./cmd/greenlync-api-gateway cmd/*.go
	@sudo docker build -t greenlync-api-gateway:dev .
	@echo "Configure your deployment registry for push" 

.PHONY: deploy-prod
deploy-prod:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o ./cmd/greenlync-api-gateway cmd/*.go
	@sudo docker build -t greenlync-api-gateway:prod .
	@echo "Configure your deployment registry for push" 

	
