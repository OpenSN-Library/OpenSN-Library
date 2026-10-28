PYTHON ?= python3

test:
	PYTHONPATH=.. $(PYTHON) -m unittest opensn.tests.test_units opensn.tests.test_schema opensn.tests.test_write opensn.tests.test_provisioning opensn.tests.test_flow opensn.tests.test_store opensn.tests.test_editing -v

check:
	$(PYTHON) opensn/tools/check_contract.py

probe:
	$(PYTHON) opensn/tools/probe_etcd_api.py

.PHONY: test check probe
