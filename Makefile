# DOCKER TASKS

DOCKER_IMAGE_NAME := budgetbot
DOCKER_IMAGE_TAG := iamhalje
DOCKER_CONTAINER_NAME := budgetbot

.PHONY: all build run extract clean

all: build run extract clean

build:
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

run:
	docker run --name $(DOCKER_CONTAINER_NAME) $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

extract:
	docker cp $(DOCKER_CONTAINER_NAME):/app/cmd/bot/budgetbot ./budgetbot

clean:
	docker stop $(DOCKER_CONTAINER_NAME) || true
	docker rm $(DOCKER_CONTAINER_NAME) || true
