build:
	go build .

docker:
	docker-compose up

run: build
	docker-compose up
	./cache_stats

goreman: build
	goreman start