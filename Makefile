tag=latest

all: server

server:
	go build -o bin/basicproject_go main.go

run:
	go run main.go

test:
	go test -v ./...

linux:
	env GOOS=linux GOARCH=amd64 go build -o bin/basicproject_go.linux main.go

dockerbuild:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-s' -o bin/basicproject_go.linux main.go

docker: dockerbuild
	docker build --platform linux/amd64 -t kobums/basicproject_go:$(tag) .

dockerrun:
	docker run --env-file .env --platform linux/amd64 -d --name="basicproject_go" -p 8007:8007 kobums/basicproject_go:$(tag)

push: docker
	docker push kobums/basicproject_go:$(tag)

clean:
	rm -f bin/basicproject_go