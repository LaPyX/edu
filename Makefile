up:
	@COMPOSE_HTTP_TIMEOUT=200 docker-compose up -d;

down:
	@COMPOSE_HTTP_TIMEOUT=200 docker-compose stop;

restart:
	make -s down
	make -s up

exec:
	docker-compose exec app sh;