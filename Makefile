
.PHONY: dockerise
dockerise:
	docker build -t instrumentation .

.PHONY: docker-compose
docker-compose:
	docker compose up
