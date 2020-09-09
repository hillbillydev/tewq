AWS_BRANCH ?= "dev"
UNAME := $(shell uname)

target:
	$(info ${HELP_MESSAGE})
	@exit 0

init:
	$(info [*] Bootstrapping CI system...)

test:
	go test ./...

build-erd:
	$(info [*] Bulilding Entity Relationship Diagram...)
	cat erd.er | docker run --rm -i kaishuu0123/erd-go | docker run --rm -i risaacson/graphviz dot -T png > erd.png

preview-erd: build-erd
ifeq ($(UNAME), Darwin)
	open erd.png
endif

define HELP_MESSAGE
Environment variables:

These variables are automatically filled at CI time.

AWS_BRANCH: "dev"
	Description: Feature branch name used as part of stacks name;

Common usage:

...::: Bootstraps environment.
$ make init

...::: Tests the whole code base.
$ make test

...::: Build Relationship Diagram
$ make build-erd
endef

