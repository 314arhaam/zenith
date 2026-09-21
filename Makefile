# Convenience only. Native Windows does not need Make; use py -3 scripts/dev.py.
PYTHON ?= python3
.PHONY: build test test-race vet fmt fmt-check py-test e2e verify dist bench
build:
	$(PYTHON) scripts/dev.py build
test:
	$(PYTHON) scripts/dev.py test
test-race:
	$(PYTHON) scripts/dev.py test --race
vet:
	$(PYTHON) scripts/dev.py vet
fmt:
	$(PYTHON) scripts/dev.py fmt
fmt-check:
	$(PYTHON) scripts/dev.py fmt --check
py-test:
	$(PYTHON) scripts/dev.py py-test
e2e:
	$(PYTHON) scripts/dev.py e2e
verify:
	$(PYTHON) scripts/dev.py verify
dist:
	$(PYTHON) scripts/dev.py dist --all
bench:
	$(PYTHON) scripts/dev.py bench
