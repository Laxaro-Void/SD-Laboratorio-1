.PHONY: run-rabbitmq stop

build-Lester:
	sudo docker compose build lester

build-Michael:
	sudo docker compose build michael

run-rabbitmq:
	sudo docker compose up -d rabbitmq
	
run-Lester:
	sudo docker compose run --rm lester

run-Michael:
	sudo docker compose run --rm michael

stop:
	-sudo docker stop michael
	-sudo docker stop lester
	-sudo docker stop rabbitmq