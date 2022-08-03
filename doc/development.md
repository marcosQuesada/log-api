# Development Notes

Development path started defining ImmuDB repository, once achieved all requirements it gets wrapped by the gRPC service layer, which is binded to the transport.
First iteration uses a gRPC server on the required API, as we are using GoogleAPI proto buffer types we can load them externally in protoc compiling step, but this will require more local component installation, an alternative is to copy/paste those definitions in the google folder, I've used this workaround to speed up the development flow.

### Proto compile process

log proto definitions, compile and autogenerate:
```
 protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    internal/proto/v1/log.proto
```

There's an extra mile that we can do to bind the grpc server to http through the gRPC Gateway implementation (it creates a gRPC client binded to the http handler layer). To allow grpc-gateway generation we need to decorate gRPC definitions, adding http descriptions on the exposed endpoints.

Using gRPC gateway plugin compilation:
```
protoc -I . --grpc-gateway_out=logtostderr=true:. internal/proto/v1/log.proto
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

### Run Application from golang binary
Get dependencies
```
go mod vendor
```
```
go build -o api
```
Before Starting the server ImmuDB needs to be started:
```
docker run -it -d --net immudb-net -p 3322:3322 -p 9497:9497 --name immudb codenotary/immudb:latest
```
Start server locally:
```
./api server
```
### Run Application from docker
Build docker image:
```
docker build -t log-api .
```
Create a bridged docker network:
```
docker network create immudb-net
```
Run Immudb from docker in the bridged network:
```
docker run -it -d --net immudb-net -p 3322:3322 -p 9497:9497 --name immudb codenotary/immudb:latest
```
Run Log-API server as:
```
docker run -it -d --net immudb-net -p 9000:9000 -p 9090:9090 -e immudb-host=immudb --name log_api log-api:latest ./app/api server 
```

### gRPC CLI Client
Cobra package has been used to generate a quick CLI command access that allows to fire all service commands.

#### Login
```
./api client login                                   
login called
2022/08/02 17:16:08 User: token:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmluY2lwYWxfaWQiOiI4ZGJmZTYyYi05YjA4LTQzYTQtOWU0My03ZTNkZTUxMzJjMDkiLCJlbWFpbCI6ImZha2VfdXNlciIsImV4cCI6MTY1OTUzOTc2OCwianRpIjoiOGRiZmU2MmItOWIwOC00M2E0LTllNDMtN2UzZGU1MTMyYzA5MTY1OTQ1MzM2ODA2NDQwNTA4MiIsImlhdCI6MTY1OTQ1MzM2OCwiaXNzIjoiTG9nIEFQSSIsInN1YiI6IkxvZ2dlciJ9.0gvDS5YZyqOt2V4NEVCTxObIrr6RTcsG4Wm-5HDcMGE"
```

#### Export JWT token to use it from the rest of commands
```
export JWT=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmluY2lwYWxfaWQiOiI4ZGJmZTYyYi05YjA4LTQzYTQtOWU0My03ZTNkZTUxMzJjMDkiLCJlbWFpbCI6ImZha2VfdXNlciIsImV4cCI6MTY1OTUzOTc2OCwianRpIjoiOGRiZmU2MmItOWIwOC00M2E0LTllNDMtN2UzZGU1MTMyYzA5MTY1OTQ1MzM2ODA2NDQwNTA4MiIsImlhdCI6MTY1OTQ1MzM2OCwiaXNzIjoiTG9nIEFQSSIsInN1YiI6IkxvZ2dlciJ9.0gvDS5YZyqOt2V4NEVCTxObIrr6RTcsG4Wm-5HDcMGE
```

#### Add Single Log Line
```
./api client add --token=$JWT --line-data=`{"source":"fake_source_a","bucket":"fake_bucket","value":"fake data value xxx","created_at":{"seconds":1659469226,"nanos":165084420}}`

2022/08/02 21:48:22 Created Log Line key fake_source_a_1659469226165084420
```

#### Add Bach of log lines
```
./api client batch --token=$JWT --lines-data=`{"lines":[{"source":"fake_source_a","bucket":"fake_bucket","value":"fake data value xxx","created_at":{"seconds":1659469108,"nanos":710408961}},{"source":"fake_source_b","bucket":"fake_bucket","value":"fake data value xaxaxax","created_at":{"seconds":1659469108,"nanos":710409242}}]}`

2022/08/02 23:22:55 Created Log Lines with keys: [fake_source_a_1659469108710408961 fake_source_b_1659469108710409242]
```

#### All log lines history
```
./api client history-all --token=$JWT
historyAll called
2022/08/02 18:10:25 History Key foo_bar_key_1_1659447503707097981 Revision [tx:3 value:"fake log content" revision:1]
2022/08/02 18:10:25 History Key foo_bar_key_1_1659447602626380233 Revision [tx:4 value:"fake log content" revision:1]
2022/08/02 18:10:25 History Key fox_bar_key_1_1659453982695309812 Revision [tx:5 value:"fake log content" revision:1]
2022/08/02 18:10:25 History Key log_size Revision [] // @TODO: Remove it ¿?
```

#### Last N transactioned log lines
```
./api client history-n --token=$JWT --number=3
history N last transactions called
2022/08/02 18:12:29 History Key fox_bar_key_1_1659453982695309812 Revision [tx:5  value:"fake log content"  revision:1]
2022/08/02 18:12:29 History Key foo_bar_key_1_1659447602626380233 Revision [tx:4  value:"fake log content"  revision:1]
2022/08/02 18:12:29 History Key foo_bar_key_1_1659447503707097981 Revision [tx:3  value:"fake log content"  revision:1]
```

#### Count Log Lines entries
```
./api client count --token=$JWT                                     

2022/08/02 17:26:34 User: total:3

```

#### List log lines by Key
```
./api client get-by-key --token=$JWT --key=foo_bar_key_1_1659447602626380233

getByKey called
2022/08/02 18:13:57 User: key:"foo_bar_key_1_1659447602626380233"  value:"fake log content"
```

#### List log lines by Key prefix

```
./api client get-by-prefix --token=$JWT --prefix=foo
getByPrefix called  with prefix  foo
2022/08/02 15:49:10 LogLine with key foo_bar_key_1_1659447503707097981: key:"foo_bar_key_1_1659447503707097981" value:"fake log content"
2022/08/02 15:49:10 LogLine with key foo_bar_key_1_1659447602626380233: key:"foo_bar_key_1_1659447602626380233" value:"fake log content"
```

#### Log Lines By Bucket (from Zset)
```
go run main.go client get-by-bucket --token=$JWT --bucket=fake_bucket


getByBucketCmd called  with bucket  fake_bucket
2022/08/03 11:36:35 LogLine with key fake_source_a_1659469108710408961: key:"fake_source_a_1659469108710408961"  value:"fake data value xxx"
2022/08/03 11:36:35 LogLine with key fake_source_b_1659469108710409242: key:"fake_source_b_1659469108710409242"  value:"fake data value xaxaxax"

```

## HTTP bindings

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

#### Store Single Log Line
```
curl -X POST -H "Authorization: Bearer $JWT" http://localhost:9090/api/v1/log -d '{"source":"fake_source","bucket":"FakeBucket-XXX","value":"FakeData-XXX","created_at":"2022-07-30T15:51:37Z"}'

{"key":"fake_source_1659196297000000000"}
```

#### Store Batch of Log Lines
```
curl -X POST -H "Authorization: Bearer $JWT" http://localhost:9090/api/v1/logs -d '{"log_lines":[{"source":fake_source", "bucket":"FakeBucket","value":"FakeData","created_at":"2022-07-30T15:51:34Z"},{"source":"fake_source_a","bucket":"FakeBucket1","value":"FakeData1","created_at":"2022-07-30T15:51:34Z"}]}'

@TODO
```

Print history of stored logs (all, last x)
```
curl -X GET -H "Authorization: Bearer $JWT" http://localhost:9090/api/v1/log/history/all   
{"histories":[{"key":"kerfel_1659369412341169159","revision":[{"tx":"3","value":"foocxx barxxxc data raw log","revision":"1"}]},{"key":"kerfel_1659369502033063593","revision":[{"tx":"4","value":"foocxx barxxxc data raw log","revision":"1"}]},{"key":"log_size","revision":[{"tx":"2","value":"\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000","revision":"1"},{"tx":"3","value":"\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0001","revision":"2"},{"tx":"4","value":"\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0002","revision":"3"}]}]}
```

Print number of stored logs
```
curl -X GET h-H "Authorization: Bearer $JWT" ttp://localhost:9090/api/v1/logs/count                                                                                             
{"total":199}
```

Get Single Log Line (Extra)
```
curl -X GET -H "Authorization: Bearer $JWT" http://localhost:9090/api/v1/log/key/kerfel_1659368898026408598                                
{"key":"kerfel_1659368898026408598","value":"foocxx barxxxc data raw log"}
```

Use JWT from http
```
curl -X GET -H "Authorization: Bearer $JWT" http://localhost:9090/api/v1/logs/count                                                                              

{"total":10}
```

Get Log Lines By Bucket
```
curl -X GET -H "Authorization: Bearer $JWT" http://localhost:9090/api/v1/log/bucket/fake_bucket       
{"log_lines":[{"key":"fake_source_a_1659469108710408961","value":"fake data value xxx"},{"key":"fake_source_b_1659469108710409242","value":"fake data value xaxaxax"}]}
```
