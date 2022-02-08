snapshot:
	goreleaser --snapshot --rm-dist

release:
	goreleaser --rm-dist

run:
	go run main.go