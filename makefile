.PHONY: build-producer build-consumer run-producer run-consumer run-rabbitmq stop

build-producer:
	sudo docker build -t producer ./producer

build-consumer:
	sudo docker build -t consumer ./consumer

run-rabbitmq:
	sudo docker run --it --name rabbitmq --rm -p 5672:5672 -p 15672:15672 rabbitmq:4-management

run-producer:
	sudo docker run --name producer --rm --link rabbitmq:rabbitmq producer

run-consumer:
	sudo docker run --name consumer --rm --link rabbitmq:rabbitmq consumer

stop:
	-sudo docker stop producer
	-sudo docker stop consumer
	-sudo docker stop rabbitmq