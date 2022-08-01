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

## HTTP bindings

Store Single Log Line
```
curl -X POST http://localhost:9090/api/v1/log -d '{"source":"fake_source","bucket":"FakeBucket-XXX","value":"FakeData-XXX","created_at":"2022-07-30T15:51:37Z"}'

{"key":"fake_source_1659196297000000000"}
```

Store Batch of Log Lines
```
curl -X POST http://localhost:9090/api/v1/logs -d '{"log_lines":[{"source":fake_source", "bucket":"FakeBucket","value":"FakeData","created_at":"2022-07-30T15:51:34Z"},{"source":"fake_source_a","bucket":"FakeBucket1","value":"FakeData1","created_at":"2022-07-30T15:51:34Z"}]}'

@TODO
```

Print history of stored logs (all, last x)
```
curl -X GET http://localhost:9090/api/v1/log/history/all   
{"histories":[{"key":"kerfel_1659369412341169159","revision":[{"tx":"3","value":"foocxx barxxxc data raw log","revision":"1"}]},{"key":"kerfel_1659369502033063593","revision":[{"tx":"4","value":"foocxx barxxxc data raw log","revision":"1"}]},{"key":"log_size","revision":[{"tx":"2","value":"\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000","revision":"1"},{"tx":"3","value":"\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0001","revision":"2"},{"tx":"4","value":"\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0002","revision":"3"}]}]}
```

Print number of stored logs
```
curl -X GET http://localhost:9090/api/v1/logs/count                                                                                             
{"total":199}
```

Get Single Log Line (Extra)
```
curl -X GET http://localhost:9090/api/v1/log/key/kerfel_1659368898026408598                                
{"key":"kerfel_1659368898026408598","value":"foocxx barxxxc data raw log"}
```
