.PHONY: all build clean deploy test run

up.local:
	@echo "[sapo-server]: up"
	docker-compose up

down.local:
	@echo "[sapo-server]: down"
	docker-compose down
