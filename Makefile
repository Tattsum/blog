# 個人ブログモノレポ用 Makefile
# GCP 用ターゲットは GCP_PROJECT_ID, REGION を環境変数または .env で設定すること

GCP_PROJECT_ID ?= kano-blog-prod
REGION         ?= asia-northeast1
IMAGE_NAME     ?= blog-api
IMAGE_TAG      ?= latest
IMAGE          := $(REGION)-docker.pkg.dev/$(GCP_PROJECT_ID)/blog-repo/$(IMAGE_NAME):$(IMAGE_TAG)

.PHONY: help docker-build-api docker-push-api docker-api proto lint test migrate-up migrate-down \
	terraform-init terraform-plan terraform-apply

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Docker (Cloud Run 用・linux/amd64):"
	@echo "  docker-build-api   - コンテナイメージをビルド (GCP_PROJECT_ID, REGION)"
	@echo "  docker-push-api   - Artifact Registry に push"
	@echo "  docker-api        - ビルド + push"
	@echo ""
	@echo "Proto / Lint / Test:"
	@echo "  proto             - buf generate (Go + TS)"
	@echo "  lint              - Markdown + Proto + Go の lint"
	@echo "  test              - go test ./..."
	@echo ""
	@echo "DB マイグレーション (DATABASE_DSN を環境変数で指定):"
	@echo "  migrate-up        - migrate up"
	@echo "  migrate-down      - migrate down"
	@echo ""
	@echo "Terraform:"
	@echo "  terraform-init    - terraform init"
	@echo "  terraform-plan    - terraform plan"
	@echo "  terraform-apply   - terraform apply"

# --- Docker (Cloud Run 用・必ず linux/amd64 でビルド) ---
docker-build-api:
	docker build --platform linux/amd64 -t $(IMAGE) -f backend/Dockerfile .

docker-push-api:
	gcloud auth configure-docker $(REGION)-docker.pkg.dev --quiet
	docker push $(IMAGE)

docker-api: docker-build-api docker-push-api

# --- Proto / Lint / Test ---
proto:
	npm run generate:proto

lint:
	npm run lint:md
	buf lint
	npm run lint:go

test:
	go test ./...

# --- マイグレーション (DATABASE_DSN を export して実行) ---
migrate-up:
	migrate -path backend/db/migrations -database "$${DATABASE_DSN}" up

migrate-down:
	migrate -path backend/db/migrations -database "$${DATABASE_DSN}" down

# --- Terraform ---
terraform-init:
	cd terraform && terraform init

terraform-plan:
	cd terraform && terraform plan

terraform-apply:
	cd terraform && terraform apply
