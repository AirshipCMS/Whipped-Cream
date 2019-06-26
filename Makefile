certs:
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 -subj "/C=US/ST=Washington/L=Seattle/O=Purely Functional/CN=wc.purelyfunctional.co" \-keyout certs/wc.key -out certs/wc.crt

build:
	go build -o bin/Whipped-Cream .

build-linux64:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -installsuffix 'static' -o bin/Whipped-Cream .

run:
	env CERT_PATH=certs/wc.crt CERT_KEY_PATH=certs/wc.key ./bin/Whipped-Cream

dev: build run

.PHONY: certs build build-linux64 run dev
