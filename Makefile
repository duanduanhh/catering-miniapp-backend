.PHONY: init
init:
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: bootstrap
bootstrap:
	cd ./deploy/docker-compose && docker compose up -d && cd ../../
	go run ./cmd/migration
	nunu run ./cmd/server

.PHONY: mock
mock:
	mockgen -source=internal/service/user.go -destination test/mocks/service/user.go
	mockgen -source=internal/repository/user.go -destination test/mocks/repository/user.go
	mockgen -source=internal/repository/repository.go -destination test/mocks/repository/repository.go

.PHONY: test
test:
	go test -coverpkg=./internal/handler,./internal/service,./internal/repository -coverprofile=./coverage.out ./test/server/...
	go tool cover -html=./coverage.out -o coverage.html

.PHONY: build
build:
	go build -ldflags="-s -w" -o ./bin/server ./cmd/server

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/server ./cmd/server

.PHONY: docker
SWR_REGISTRY ?= swr.cn-north-1.myhuaweicloud.com
SWR_ORG      ?= catering-cyxx
IMAGE_NAME   ?= miniapp-backend
IMAGE_TAG    ?= v1-20250114.1

docker:
	docker buildx build --platform linux/amd64 --pull=false --load -f deploy/build/Dockerfile \
		--build-arg APP_RELATIVE_PATH=./cmd/server \
		-t $(SWR_REGISTRY)/$(SWR_ORG)/$(IMAGE_NAME):$(IMAGE_TAG) \
		--push .

.PHONY: swag
swag:
	swag init  -g cmd/server/main.go -o ./docs
