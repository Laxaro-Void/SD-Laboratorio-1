# Makefile #
build-protoc:
	protoc --go_out=Akatsuki/proto --go-grpc_out=Akatsuki/proto Akatsuki/proto/message.proto
	protoc --go_out=Anbu/proto --go-grpc_out=Anbu/proto Anbu/proto/message.proto
	protoc --go_out=EquiposNinja/proto --go-grpc_out=EquiposNinja/proto EquiposNinja/proto/message.proto
	protoc --go_out=Hokage/proto --go-grpc_out=Hokage/proto Hokage/proto/message.proto

## Production
docker-akatsuki:
	sudo docker-compose build akatsuki
	sudo docker-compose run akatsuki

docker-anbu:
	sudo docker-compose build anbu
	sudo docker-compose run anbu

docker-equiposninja:
	sudo docker-compose build equiposninja
	sudo docker-compose run equiposninja

docker-hokage:
	sudo docker-compose build hokage
	sudo docker-compose run hokage

stop-docker-akatsuki:
	sudo docker-compose stop akatsuki

stop-docker-anbu:
	sudo docker-compose stop anbu

stop-docker-hokage:
	sudo docker-compose stop hokage

stop-docker-equiposninja:
	sudo docker-compose stop equiposninja

## Localhost
local-docker-akatsuki:
	sudo docker-compose -f compose.localhost.yaml build akatsuki
	sudo docker-compose -f compose.localhost.yaml run akatsuki

local-docker-anbu:
	sudo docker-compose -f compose.localhost.yaml build anbu
	sudo docker-compose -f compose.localhost.yaml run anbu

local-docker-equiposninja:
	sudo docker-compose -f compose.localhost.yaml build equiposninja
	sudo docker-compose -f compose.localhost.yaml run equiposninja

local-docker-hokage:
	sudo docker-compose -f compose.localhost.yaml build hokage
	sudo docker-compose -f compose.localhost.yaml run hokage

stop-local-docker-akatsuki:
	sudo docker-compose -f compose.localhost.yaml stop akatsuki

stop-local-docker-anbu:
	sudo docker-compose -f compose.localhost.yaml stop anbu

stop-local-docker-hokage:
	sudo docker-compose -f compose.localhost.yaml stop hokage

stop-local-docker-equiposninja:
	sudo docker-compose -f compose.localhost.yaml stop equiposninja