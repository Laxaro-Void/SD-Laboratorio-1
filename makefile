.PHONY: build-Lester build-Michael build-Francis build-Trevor build-rabbitmq run-rabbitmq run-Lester run-Michael run-Francis run-Trevor stop

build-Lester:
	sudo docker-compose build lester

build-Michael:
	sudo docker-compose build michael

build-Francis:
	sudo docker-compose build francis

build-Trevor:
	sudo docker-compose build trevor

build-rabbitmq:
	sudo docker-compose build rabbitmq

run-rabbitmq:
	sudo docker-compose up -d rabbitmq
	
run-Lester:
	sudo docker-compose run --rm lester

run-Michael:
	sudo docker-compose run --rm michael

run-Francis:
	sudo docker-compose run --rm francis

run-Trevor:
	sudo docker-compose run --rm trevor

stop:
	-sudo docker stop trevor
	-sudo docker stop francis
	-sudo docker stop michael
	-sudo docker stop lester
	-sudo docker stop rabbitmq