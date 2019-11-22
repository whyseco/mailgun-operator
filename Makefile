#!/usr/bin/make -f

all: bundle.yaml

bundle.yaml: deploy/operator.yaml deploy/role_binding.yaml deploy/role.yaml deploy/service_account.yaml
		awk 'FNR==1{print "---"}1' deploy/operator.yaml deploy/role_binding.yaml deploy/role.yaml deploy/service_account.yaml | awk '{if (NR!=1) {print}}' > deploy/bundle.yaml