#!/usr/bin/make -f

all: bundle.yaml

bundle.yaml: deploy/operator.yaml deploy/role_binding.yaml deploy/role.yaml deploy/service_account.yaml
		cat deploy/operator.yaml deploy/role_binding.yaml deploy/role.yaml deploy/service_account.yaml > deploy/bundle.yaml