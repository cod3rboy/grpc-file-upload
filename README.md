# GRPC File Upload

This application demonstrates client streaming and bidirectional streaming modes in gRPC.
Server exposes RPC functions for uploading a file. Client uploads file by invoking these server functions.

RPC functions -

- **UploadFile:** Upload file using client streaming.
- **UploadFileWithProgress:** Upload file using client streaming and reports progress using server streaming.

## Generating code from proto files

`protoc --proto_path="./proto" --go_out="./gen/uploader" --go-grpc_out="./gen/uploader" --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative uploader.proto`

## Running server

`go run ./cmd/server <flags>`

### CLI flags

| Flag | Description                           | Usage                | Default Value |
| ---- | ------------------------------------- | -------------------- | ------------- |
| port | server port                           | `-port <value>`      | 8000          |
| dir  | directory which stores uploaded files | `-dir <path-to-dir>` | tmp           |

## Running client

`go run ./cmd/client <flags>`

### CLI flags

| Flag              | Description                | Usage                     | Default Value  |
| ----------------- | -------------------------- | ------------------------- | -------------- |
| file (_required_) | file to upload to server   | `-file <path-to-file>`    | _none_         |
| server            | server address (host:port) | `-server <host:port>`     | localhost:8000 |
| progress          | display upload progress    | `-progress <true\|false>` | false          |
