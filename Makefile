test: fmt vet
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

run: fmt vet
	go run ./main.go