# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=tarantulago
MAIN_PATH=./cmd/bot/main.go
SRC_DIR=./src

# Docker parameters
DOCKER_COMPOSE=docker-compose
DOCKER_IMAGE_NAME=tarantulago

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build test clean run deps docker-build docker-run help lint check

all: clean deps build

build:
	cd $(SRC_DIR) && $(GOBUILD) $(LDFLAGS) -o ../$(BINARY_NAME) $(MAIN_PATH)

test:
	cd $(SRC_DIR) && $(GOTEST) -v ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out

run:
	cd $(SRC_DIR) && $(GOBUILD) -o ../$(BINARY_NAME) $(MAIN_PATH)
	./$(BINARY_NAME)

deps:
	cd $(SRC_DIR) && $(GOMOD) download
	cd $(SRC_DIR) && $(GOMOD) tidy

# Development tools installation
install-tools:
	$(GOGET) -u golang.org/x/lint/golint
	$(GOGET) -u github.com/kisielk/errcheck
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint

# Linting and code quality
lint:
	cd $(SRC_DIR) && golangci-lint run ./...
	cd $(SRC_DIR) && golint ./...
	cd $(SRC_DIR) && errcheck ./...

# Test with coverage
test-coverage:
	cd $(SRC_DIR) && $(GOTEST) -coverprofile=../coverage.out ./...
	go tool cover -html=coverage.out

# Docker commands
docker-build:
	docker build -t $(DOCKER_IMAGE_NAME) .

docker-run:
	docker run --env-file .env $(DOCKER_IMAGE_NAME)

docker-compose-up:
	$(DOCKER_COMPOSE) up -d

docker-compose-down:
	$(DOCKER_COMPOSE) down

# Database migrations (placeholder - implement according to your migration tool)
migrate-up:
	@echo "Running database migrations up"
	# Add your migration command here

migrate-down:
	@echo "Rolling back database migrations"
	# Add your migration command here

# Development mode with hot reload (requires air: https://github.com/cosmtrek/air)
dev:
	cd $(SRC_DIR) && air

# Check for required tools and environment variables
check:
	@echo "Checking required tools and environment variables..."
	@which go > /dev/null || (echo "Error: go is not installed" && exit 1)
	@test -n "$$TELEGRAM_BOT_TOKEN" || (echo "Error: TELEGRAM_BOT_TOKEN is not set" && exit 1)
	@test -n "$$POSTGRES_URL" || (echo "Error: POSTGRES_URL is not set" && exit 1)

# Production build with optimizations
prod-build:
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a $(LDFLAGS) -o ../$(BINARY_NAME) $(MAIN_PATH)

# Help command
help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build files"
	@echo "  make run           - Build and run the application"
	@echo "  make deps          - Download and tidy dependencies"
	@echo "  make lint          - Run linters"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run Docker container"
	@echo "  make dev           - Run in development mode with hot reload"
	@echo "  make check         - Check requirements"
	@echo "  make prod-build    - Build for production"