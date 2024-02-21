generate:
	mkdir -p ./gen/uploader && \
	protoc \
	--proto_path="./proto" \
	--go_out="./gen/uploader" \
	--go-grpc_out="./gen/uploader" \
	--go_opt=paths=source_relative \
	--go-grpc_opt=paths=source_relative \
	uploader.proto