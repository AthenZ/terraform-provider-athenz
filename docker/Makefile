DOCKER_NETWORK := $(if $(DOCKER_NETWORK),$(DOCKER_NETWORK),athenz)
BASE_DIR=$(shell git rev-parse --show-toplevel)

deploy-local:
	# generate self-signed certificates
	docker run --rm -t -v "$(BASE_DIR):/terraform-provider-athenz" --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh /terraform-provider-athenz/docker/setup-scripts/self-signed-certificates.sh
	# ZMS
	docker run --rm -t -v "$(BASE_DIR):/terraform-provider-athenz" --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh /terraform-provider-athenz/docker/setup-scripts/zms-auto-config.sh
	sh "./deploy-scripts/zms-deploy-local.sh"
	docker run --rm -t --network="$(DOCKER_NETWORK)" -v "$(BASE_DIR):/terraform-provider-athenz" --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh /terraform-provider-athenz/docker/deploy-scripts/zms-debug.sh

CONTAINERS := $(shell docker ps -aq --filter 'label=org.label-schema.url=https://www.athenz.io/')
remove-all: remove-containers remove-networks remove-files
remove-containers:
ifneq ($(CONTAINERS),)
	docker stop $(shell docker ps -aq --filter 'label=org.label-schema.url=https://www.athenz.io/')
	docker rm $(shell docker ps -aq --filter 'label=org.label-schema.url=https://www.athenz.io/')
else
	$(info No containers to be removed.)
endif
remove-networks:
	docker network rm $(DOCKER_NETWORK) || true

remove-files:
	rm -rf ./prod
	rm -rf ./logs
	rm -rf ./jars
	rm -rf ./zts
	rm -rf ./ui
	rm -rf ./zms/var
	
clean: remove-all
	docker image rm athenz/athenz-setup-env || true
	docker image rm athenz/athenz-zms-db || true
	docker image rm athenz/athenz-zms-server || true
