DOCKER_NETWORK := $(if $(DOCKER_NETWORK),$(DOCKER_NETWORK),athenz)
DOCKER_DNS := $(if $(DOCKER_DNS),$(DOCKER_DNS),8.8.8.8)
LOCAL_ENV_NS := $(shell ./local-nameserver.sh)

ifeq ($(SD_SOURCE_DIR),)
	SD_SOURCE_DIR=$(shell git rev-parse --show-toplevel)
endif

deploy:
	cp -R $(SD_SOURCE_DIR) $(SD_DIND_SHARE_PATH)/
	find $(SD_DIND_SHARE_PATH)

	# generate self-signed certificates
	docker run --rm -it -v "$(SD_DIND_SHARE_PATH)/terraform-provider-athenz:/terraform-provider-athenz" athenz/athenz-setup-env sh -c 'set -x ; /terraform-provider-athenz/docker/setup-scripts/self-signed-certificates.sh'
	# ZMS
	docker run --rm -it -v "$(SD_DIND_SHARE_PATH)/terraform-provider-athenz:/terraform-provider-athenz" athenz/athenz-setup-env sh -c 'set -x ; /terraform-provider-athenz/docker/setup-scripts/zms-auto-config.sh'
	sh "./deploy-scripts/zms-deploy.sh"
	docker run --rm -it --network="$(DOCKER_NETWORK)" -v "$(SD_DIND_SHARE_PATH)/terraform-provider-athenz:/terraform-provider-athenz" athenz/athenz-setup-env sh -c 'set -x ; /terraform-provider-athenz/docker/deploy-scripts/zms-debug.sh'

deploy-local:
	cp -R $(SD_SOURCE_DIR) $(SD_DIND_SHARE_PATH)/
	find $(SD_DIND_SHARE_PATH)

	# generate self-signed certificates
	docker run --rm -it -v "$(SD_DIND_SHARE_PATH)/terraform-provider-athenz:/terraform-provider-athenz" athenz/athenz-setup-env sh -c 'set -x ; /terraform-provider-athenz/docker/setup-scripts/self-signed-certificates.sh'
	# ZMS
	docker run --rm -it -v "$(SD_DIND_SHARE_PATH)/terraform-provider-athenz:/terraform-provider-athenz" athenz/athenz-setup-env sh -c 'set -x ; /terraform-provider-athenz/docker/setup-scripts/zms-auto-config.sh'
	sh "./deploy-scripts/zms-deploy-local.sh"
	docker run --rm -it --network="$(DOCKER_NETWORK)" -v "$(SD_DIND_SHARE_PATH)/terraform-provider-athenz:/terraform-provider-athenz" athenz/athenz-setup-env sh -c 'set -x ; /terraform-provider-athenz/docker/deploy-scripts/zms-debug.sh'


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
