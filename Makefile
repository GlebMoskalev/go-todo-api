CONFIG_FILE ?= ./config/local.yaml

ENV := $(shell yq '.env' $(CONFIG_FILE))
DB_HOST = $(shell yq '.database.host' $(CONFIG_FILE))
DB_PORT = $(shell yq '.database.port' $(CONFIG_FILE))
DB_NAME = $(shell yq '.database.name' $(CONFIG_FILE))
DB_USER = $(shell yq '.database.user' $(CONFIG_FILE))
DB_PASSWORD = $(shell yq '.database.password' $(CONFIG_FILE))

DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable


migration_up:
	migrate -path migrations -database $(DB_URL) up

migration_down:
	migrate -path migrations -database $(DB_URL) down