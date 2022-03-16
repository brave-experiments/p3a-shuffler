.PHONY: all test lint eif install clean

enclave_cid = 5
binary = p3a-shuffler
godeps = *.go go.mod go.sum

all: test lint $(binary)

test:
	go test -cover ./...

lint:
	golangci-lint run -E gofmt -E revive --exclude-use-default=false

image:
	$(eval image=$(shell ko publish --local . 2>/dev/null))
	@echo "Built image URI: $(image)."
	$(eval digest=$(shell echo $(image) | cut -d ':' -f 2))
	@echo "SHA-256 digest: $(digest)"

eif: image
	nitro-cli build-enclave --docker-uri $(image) --output-file ko.eif
	$(eval enclave_id=$(shell nitro-cli describe-enclaves | jq -r "if .[].EnclaveCID == $(enclave_cid) then .[].EnclaveID else \"\" end"))
	@if [ "$(enclave_id)" != "" ]; then nitro-cli terminate-enclave --enclave-id $(enclave_id); fi
	@echo "Starting enclave."
	nitro-cli run-enclave --cpu-count 2 --memory 2500 --enclave-cid $(enclave_cid) --eif-path ko.eif --debug-mode
	nitro-cli console --enclave-id \
		$$(nitro-cli describe-enclaves | jq -r "if .[].EnclaveCID == $(enclave_cid) then .[].EnclaveID else \"\" end")

$(binary): $(godeps)
	go build -o $(binary) .

install: $(godeps)
	go install

clean:
	rm $(binary)
