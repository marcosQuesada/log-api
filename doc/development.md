# Development Notes

Development path started defining ImmuDB repository, once achieved all requirements it gets wrapped by the gPC service layer, which is binded to the transport.
First iteration achieves a gRPC server on the required API. As we are using GoogleAPI types we can load them externally in protoc compiling step, but this will require more local component installation, an alternative is to copy/paste those definitions in the google folder, I've used this workaround to speed up the development flow.

## API centric development

log proto definitions, compile and autogenerate:
```
 protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    internal/proto/v1/log.proto

```

There's an extra mile that we can do to bind the grpc server to http through the gRPC Gateway implementation (it creates a gRPC client binded to the http handler layer). To allow grpc-gateway generation we need to decorate gRPC definitions, adding http descriptions on the exposed endpoints.

Log Http Gateway proto compilation:
```
protoc -I . --grpc-gateway_out=logtostderr=true:. internal/proto/v1/log.proto
```

## gRPC client
Using cobra is really useful on CLI command creation, a client single log line creation has been added, which expects 3 arguments as Source, Bucket and Value:

```
go run main.go client --token=$JWT foo_bar_key auth-app "fake log content"
```

## HTTP bindings @TODO: add auth JWT tokens on examples

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

## Authorization system

A JWT authorization system is implemented as UnaryInterceptor (StreamInterceptor has not been implemented as right now is not required)

To get JWT credentials we just need to invoke Login request , any username and password will be replied with a valid credential.

This auth scheme protects all API endpoints, except the login request one, and so, JWT auth needs to be attached on each request.
- gRPC client as Outgoing context attachment
- http client as Authorization header

### Auth Proto definition and compiling
```
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=logtostderr=true:. \
    internal/proto/v1/auth.proto
    
```

### AUTH credentials Inclusion
Login
```
curl -H "Content-Type: application/json" -H 'accept: application/json' -X POST http://localhost:9090/api/v1/auth -d '{"username":"baldur","password":"password"}'

{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmluY2lwYWxfaWQiOiI1YjkxOTdjMy1iNGRkLTQ5ODItOTM5MS03ZjZlNjhiNWY2MDYiLCJlbWFpbCI6ImJhbGR1ciIsImV4cCI6MTY1OTUyMDM4MiwianRpIjoiNWI5MTk3YzMtYjRkZC00OTgyLTkzOTEtN2Y2ZTY4YjVmNjA2MTY1OTQzMzk4MjI4NjE1MTI2NiIsImlhdCI6MTY1OTQzMzk4MiwiaXNzIjoiTG9nIEFQSSIsInN1YiI6IkxvZ2dlciJ9.vuWzqaUm6jQVEryN2kEPSDv8Zy0qQMUjD_COWriVdec"}`
```

export JWT token:
```
export JWT=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmluY2lwYWxfaWQiOiI1YjkxOTdjMy1iNGRkLTQ5ODItOTM5MS03ZjZlNjhiNWY2MDYiLCJlbWFpbCI6ImJhbGR1ciIsImV4cCI6MTY1OTUyMDM4MiwianRpIjoiNWI5MTk3YzMtYjRkZC00OTgyLTkzOTEtN2Y2ZTY4YjVmNjA2MTY1OTQzMzk4MjI4NjE1MTI2NiIsImlhdCI6MTY1OTQzMzk4MiwiaXNzIjoiTG9nIEFQSSIsInN1YiI6IkxvZ2dlciJ9.vuWzqaUm6jQVEryN2kEPSDv8Zy0qQMUjD_COWriVdec
```

Use JWT from GRPC
```
go run main.go client --token=$JWT foo_bar_key auth-app "fake log content"

2022/08/02 11:55:54 client called, arguments [foo_bar_key auth-app fake log content] 
2022/08/02 11:55:54 Created Log Line key foo_bar_key_1659434154176761432
2022/08/02 11:55:54 History Key foo_bar_key_1659434154176761432 Revision [tx:12  value:"fake log content"  revision:1]

```

Use JWT from http
```
curl -X GET -H "Authorization: Bearer $JWT" http://localhost:9090/api/v1/logs/count                                                                              

{"total":10}
```
