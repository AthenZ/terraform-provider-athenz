DOCKER_NETWORK := $(if $(DOCKER_NETWORK),$(DOCKER_NETWORK),athenz)
ifeq ($(SD_SOURCE_DIR),)
	SD_SOURCE_DIR=$(shell git rev-parse --show-toplevel)
endif

deploy-test:
	echo " SD_SOURCE_DIR is: $(SD_SOURCE_DIR)"
	# generate self-signed certificates
	cp -R $(SD_SOURCE_DIR) $(SD_DIND_SHARE_PATH)/
	find $(SD_DIND_SHARE_PATH)
	docker run --rm -t -v "$(SD_DIND_SHARE_PATH)/terraform-provider-athenz:/terraform-provider-athenz" athenz/athenz-setup-env sh -c "set -x ; find /terraform-provider-athenz"
#	docker run --rm -t -v /var/run/docker.sock:/var/run/docker.sock --mount src="$(SD_SOURCE_DIR)",target=/terraform-provider-athenz,type=bind --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh -c "set -x ; find /terraform-provider-athenz"
#	docker run --rm -t --mount src="$(SD_SOURCE_DIR)",target=/terraform-provider-athenz,type=bind --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh -c "set -x ; find /terraform-provider-athenz"
#	docker run --rm -t -v "$(SD_SOURCE_DIR):/terraform-provider-athenz" -v /var/run/docker.sock:/var/run/docker.sock --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh -c "set -x ; find /terraform-provider-athenz"
#	docker run --rm --mount type=bind,src="$(SD_SOURCE_DIR)",dst=/terraform-provider-athenz --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh -c "set -x ; find /terraform-provider-athenz"
#	docker run --rm -v /var/run/docker.sock:/var/run/docker.sock --mount type=bind,src="$(SD_SOURCE_DIR)",dst=/terraform-provider-athenz --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh -c "set -x ; find /terraform-provider-athenz"
#	docker run --rm --mount type=bind,src="/host_mnt/$(SD_SOURCE_DIR)",dst=/terraform-provider-athenz -v /var/run/docker.sock:/var/run/docker.sock --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh -c "set -x ; find /terraform-provider-athenz"

#	chmod -R 777 $(SD_SOURCE_DIR)
#	chmod -R 777 $(SD_ROOT_DIR)
#	cp -R $(SD_SOURCE_DIR) $(SD_ROOT_DIR)
#	find $(SD_ROOT_DIR)/terraform-provider-athenz 
#	docker run --rm -t --mount src="$(SD_ROOT_DIR)",target=/terraform-provider-athenz,type=bind --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh -c "set -x ; find /terraform-provider-athenz"
#	docker run --rm -t --mount src="$(SD_ROOT_DIR)",target=/terraform-provider-athenz,type=bind --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh -c "set -x ; find /"

#	docker run --rm -t -v "$(SD_ROOT_DIR)/terraform-provider-athenz:/terraform-provider-athenz" --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh -c "set -x ; find /terraform-provider-athenz"
	
deploy-local:
	# generate self-signed certificates
	docker run --rm -t -v "$(SD_SOURCE_DIR):/terraform-provider-athenz" --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh /terraform-provider-athenz/docker/setup-scripts/self-signed-certificates.sh
	# ZMS
	docker run --rm -t -v "$(SD_SOURCE_DIR):/terraform-provider-athenz" --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh /terraform-provider-athenz/docker/setup-scripts/zms-auto-config.sh
	sh "./deploy-scripts/zms-deploy-local.sh"
	docker run --rm -t --network="$(DOCKER_NETWORK)" -v "$(SD_SOURCE_DIR):/terraform-provider-athenz" --user "$(shell id -u):$(shell id -g)" athenz/athenz-setup-env sh /terraform-provider-athenz/docker/deploy-scripts/zms-debug.sh

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
