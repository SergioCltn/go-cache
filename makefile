build:
	go build -o bin/main main.go

run/server:
	cd server && go run .

run/client:
	cd client && go run .