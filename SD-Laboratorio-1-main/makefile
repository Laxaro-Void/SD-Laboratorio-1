.PHONY: build-producer build-consumer run-producer run-consumer run-rabbitmq stop

build-producer:
	sudo docker compose build producer

build-consumer:
	sudo docker compose build consumer

run-rabbitmq:
	sudo docker compose up -d rabbitmq
	
run-producer:
	sudo docker compose run --rm producer

run-consumer:
	sudo docker compose run --rm consumer
stop:
	-sudo docker stop producer
	-sudo docker stop consumer
	-sudo docker stop rabbitmq