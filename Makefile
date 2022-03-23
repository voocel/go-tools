default:gen

gen:
	protoc -I ./proto/ ./proto/*.proto --go_out=plugins=grpc:./proto/
	# protoc-go-inject-tag --input=./proto/*.pb.go

.PHONY: gentls
gentls:
	mkdir -p certs
	openssl genrsa -out certs/server.key 2048
	openssl ecparam -genkey -name secp384r1 -out certs/server.key
	openssl req -new -x509 -sha256 -key certs/server.key -out certs/server.crt -days 3650

santls:
	openssl genrsa -out server.key 2048

	openssl req -new -key server.key \
			-out server.csr \
			-subj "/C=GB/L=China/O=test/CN=www.test.com" \
			-reqexts SAN -config <(cat /etc/ssl/openssl.cnf <(printf "\n[SAN]\nsubjectAltName=DNS:*.test.com,DNS:*.test2.com"))

	openssl openssl x509 -req -days 3650 \
			-in server.csr -out server.crt \
			-CA ca.crt -CAkey ca.key -CAcreateserial \
			-extensions SAN \
			-extfile <(cat /etc/ssl/openssl.cnf <(printf "\n[SAN]\nsubjectAltName=DNS:*.test.com,DNS:*.test2.com"))
