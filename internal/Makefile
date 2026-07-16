# Makefile for distributed-kv-store

IMAGE_REPO = github.com/anvitha0403/distributed-kvstore
IMAGE_TAG ?= 0.0.10
IMAGE = $(IMAGE_REPO):$(IMAGE_TAG)

CHART_PATH = deploy/kvstore
RELEASE_NAME = kvstore

# -----------------------------
# Go build / tidy
# -----------------------------
tidy:
		go mod tidy

build:
		
		go build -o binary ./cmd/distributed-kvs

run:
		./binary

# -----------------------------
# Docker
# -----------------------------
docker-build:
# 		docker rmi github.com/anvitha0403/distributed-kvstore:0.0.1

		docker build -t $(IMAGE) .

docker-run:
		docker run --rm -it $(IMAGE) /bin/distributed_kvs

docker-shell:
		docker run --rm -it $(IMAGE) sh

docker-ls-bin:
		docker run --rm -it $(IMAGE) ls -l /bin

# -----------------------------
# Helm / Kubernetes
# -----------------------------
helm-install:
		helm install $(RELEASE_NAME) $(CHART_PATH)

helm-upgrade:
		helm upgrade $(RELEASE_NAME) $(CHART_PATH)

helm-uninstall:
		helm uninstall $(RELEASE_NAME)

helm-ls:
		helm ls

kubectl-pods:
		kubectl get pods

kubectl-describe:
		kubectl describe pod $(RELEASE_NAME)-kvstore-0

kubectl-logs:
		kubectl logs kvstore-0 -c kvstore

kubectl-exec:
		kubectl exec -it kvstore-0 -c kvstore -- /bin/bash

kubectl-cat-config:
		kubectl exec -it $(RELEASE_NAME)-kvstore-0 -c $(RELEASE_NAME)-kvstore -- cat /var/run/kvstore/config.yaml

# -----------------------------
# Debug pod
# -----------------------------
debug-pod:
		kubectl run -it debug --image=$(IMAGE) --rm --restart=Never -- sh

debug-cat-config:
		kubectl run -it debug --image=$(IMAGE) --rm --restart=Never -- 	

# -----------------------------
# Curl tests
# -----------------------------
curl-get:
		curl -i http://localhost:8080/kv?key=name

curl-get2:
		curl -i http://localhost:8080/kv?key=Name

curl-delete:
		curl -X DELETE http://localhost:8080/kv?key=name1 -vvv
