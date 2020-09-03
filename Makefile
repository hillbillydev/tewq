AWS_BRANCH ?= "dev"

target:
	$(info ${HELP_MESSAGE})
	@exit 0

init:
	$(info [*] Bootstrapping CI system...)

build-erd: # TODO FIX only works with Docker installed and on Mac...
	$(info [*] Bulilding Entity Relationship Diagram...)
	cat erd.er | docker run --rm -i kaishuu0123/erd-go | docker run --rm -i risaacson/graphviz dot -T png > erd.png
	open erd.png

define HELP_MESSAGE
Environment variables:

These variables are automatically filled at CI time.

AWS_BRANCH: "dev"
	Description: Feature branch name used as part of stacks name;

Common usage:

...::: Bootstraps environment.
$ make init

...::: Build Relationship Diagram
$ make build-erd
endef

