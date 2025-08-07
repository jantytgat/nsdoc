clean:
	sh scripts/clean.sh

snapshot:
	goreleaser release --snapshot --clean