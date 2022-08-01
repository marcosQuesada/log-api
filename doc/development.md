# Development Notes

## API centric development
proto definitions, compile and autogenerate:
```
 protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    internal/proto/v1/log.proto

```
Http Gateway
```
protoc -I . --grpc-gateway_out=logtostderr=true:. internal/proto/v1/log.proto
```