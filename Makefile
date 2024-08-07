.PHONY: migration
migration: #  example: make migration name=add-smth
	docker run --rm \
    -v $(realpath ./internal/storage/migrations):/migrations \
    migrate/migrate:v4.16.2 \
        create \
        -dir /migrations \
        -ext .sql \
        -seq -digits 3 \
        $(name)

.PHONY: lint
lint: _golangci-lint-rm-unformatted-report

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-golangci-lint run -c .golangci.yml > ./golangci-lint/report-unformatted.json

.PHONY: _golangci-lint-format-report
_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

.PHONY: _golangci-lint-rm-unformatted-report
_golangci-lint-rm-unformatted-report: _golangci-lint-format-report
	rm ./golangci-lint/report-unformatted.json

.PHONY: golangci-lint-clean
golangci-lint-clean:
	sudo rm -rf ./golangci-lint


.PHONY: mocks
mocks:
	mockgen -source=internal/storage/storage.go -destination=internal/storage/mocks/storage_mock.gen.go -package=mocks

RAWFILE:=coverage.out
HTMLREPORT:=coverage.html

.PHONY: coverage
coverage:
	go test ./internal/handlers -coverprofile=$(RAWFILE) && \
 	go tool cover -html=$(RAWFILE) -o $(HTMLREPORT)

.PHONY: tests
tests:
	cd ./internal/handlers && go test . -count 1 -v