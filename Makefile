#!/usr/bin/make -f

SRC=deploy/operator.yaml deploy/role_binding.yaml deploy/role.yaml deploy/service_account.yaml deploy/crds/mailgun.com_mailgundomains_crd.yaml deploy/crds/mailgun.com_mailgunroutes_crd.yaml deploy/crds/mailgun.com_mailgunwebhooks_crd.yaml

all: bundle.yaml

bundle.yaml: $(SRC)
		awk 'FNR==1{print "---"}1' $(SRC) | awk '{if (NR!=1) {print}}' > deploy/bundle.yaml