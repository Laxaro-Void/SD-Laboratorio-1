# Makefile #
build-protoc:
	protoc --go_out=Akatsuki/proto --go-grpc_out=Akatsuki/proto Akatsuki/proto/*.proto
	protoc --go_out=Anbu/proto --go-grpc_out=Anbu/proto Anbu/proto/*.proto
	protoc --go_out=EquiposNinja/proto --go-grpc_out=EquiposNinja/proto EquiposNinja/proto/*.proto
	protoc --go_out=Hokage/proto --go-grpc_out=Hokage/proto Hokage/proto/*.proto

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
	sudo docker-compose -f compose.localhost.yaml run --use-aliases --remove-orphans akatsuki

local-docker-anbu:
	sudo docker-compose -f compose.localhost.yaml build anbu
	sudo docker-compose -f compose.localhost.yaml run --use-aliases --remove-orphans anbu

local-docker-equiposninja:
	sudo docker-compose -f compose.localhost.yaml build equiposninja
	read -p "Cantidad de Equipos Ninja: " cantidad; \
	for i in $$(seq 1 $$cantidad); do \
		nohup alacritty -e sudo docker-compose -f compose.localhost.yaml run --use-aliases --remove-orphans equiposninja & \
	done

local-docker-hokage:
	sudo docker-compose -f compose.localhost.yaml build hokage
	sudo docker-compose -f compose.localhost.yaml run --use-aliases --remove-orphans hokage

stop-local-docker-akatsuki:
	sudo docker-compose -f compose.localhost.yaml stop akatsuki

stop-local-docker-anbu:
	sudo docker-compose -f compose.localhost.yaml stop anbu

stop-local-docker-hokage:
	sudo docker-compose -f compose.localhost.yaml stop hokage

stop-local-docker-equiposninja:
	sudo docker-compose -f compose.localhost.yaml stop equiposninja
