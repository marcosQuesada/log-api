# Log-API

The goal is to write a **simple REST or GRPC service in Golang that uses immudb as a database to store lines of logs**.
Service and immudb should be easily deployable using docker, docker-compose or similar.

There should be another **simple testing tool that allows you to easily create log entries.**

The service should allow to:
- **Store single log line**
- **Store batch of log lines**
- **Print history of stored logs (all, last x)**
- **Print number of stored logs**
- (optional) Simple authentication mechanism to restrict read/write access.
- (optional) **Support for log buckets**
    - logs from different applications can be separated
    - i.e. depending on source or some token used.

It is not required to develop a full blown solution, but we should be able to see that you know how to build services, work with databases (immudb), how to test code and how to ship it.

## Resources:
- Immudb: https://github.com/codenotary/immudb
- Immudb docs: https://docs.immudb.io/master/
- Immudb go sdk: https://github.com/codenotary/immudb/tree/master/pkg/client


## ImmuDB repository Design
Fist step has been design the immudb destination structure, AFAIK requirements points to the key/value implementation, as it solves Single & Batch LogLine insertions. (add & Sadd)
All LogLines History is achieved getting all created logLines using Scan command with an empty prefix.
Last N executed transactions can be resolved by getting ImmuDB current state, getting last Transaction ID, and then TxScan usage allows to get last N transactions. 

The corner case arrives at "Number of stored logs", my first understanding on that was using Count command, but, It seems not ready yet (I tested last master version from immudb pointing to v1.3.2-0.20220726101823-fb0428237486 version, getting same UnImplemented method).

As this is a requirement I found a compromise solution using schema preconditions, the idea is to use a unique ImmuDB key to store LogLines total number, use transactions on add & batch addition to control transactional behaviour.

The workflow goes like this:
- LogLine addition establish the precondition that logLine Key must not exist 
- In the transactional scope we get the current keySize value, I establish a secondary Precondition were the counter must be not changed concurrently
  - This precondition will fail on concurrent data insertion, right now is a trade-off solution, it needs more work to do to implement a full solution
  - A consolidated solution will use a Compare and Swap approach, and so, a retry loop is required on precondition fail
    - planned solution with retry loop will be eventual consistent, and it will reduce the overall system performance indeed.

#### Key Composition:
I defined a log with a Source and a value that happens on T timestamp on nanos, (avoid key overlapping is the overall goal here), key is obtained from concatenation:
```
source_ts( as nanos)
example: foo_bar_key_1659437280251549814
```

This enables us to use Scan with prefixes to get all related LogLines from the same source.

#### Log Bucket Support
Log Bucket can be concatenated in the key names, and so using Scan will enable us to filter all log lines by bucket, this will result in this key composition:
```
bucket_source_ts( as nanos)
example: payment_user_log_1659437280251549814
```

Secondary Indexes through ZSET seems a more effective solution to achieve Support for log buckets.
On this implementation: 
- ZSets defined by bucket
- LogLine key is added to the destination sorted set

This enables a better trade off (IMHO), as we still have the key composition benefits trough Scan command, filtering log lines by Source is enabled, and using Zset we can separate logs from different applications using different buckets

The whole application has been designed as an API centric application, focused on gRPC proto definition and heavy usage of protoc compilers extra tools.

Proposed solution uses:
- gRPC server 
- http server trough gRPC Gateway extension
- auth scheme implemented with JWT tokens as gRPC unary interceptor (stream interceptor is not implemented as no streams are defined yet)
  - auth API as been added to get jwt credentials over a faked User Repository (any user & password will succeed)
  - all API endpoints are restricted with the Login exception
  - gRPC and http client need to handle manually JWT token inclusion
    - client side interceptors seems the way to go to achieve full generation & renovation in a transparent manner

More info about the development path can be found in the path: /doc/development.md

## Assumptions
- key/value mode does not offer Count command yet
- SQL mode does not offer detailed histories as required

## Improvements
- grpc client side auth interceptor
- next steps....
  - Precondition fail on concurrent writers needs a retry loop
- errorGroups with context to handle grpc and http servers


## TODO
- Add OS environment var bindings
  - immudb-host ISSUE
- server entryPoint
  - use errorGroup with context on both transport server
  - handle application shutdown in a controlled manner 
    - use SigTerm and SigKill signals

- e2e test
  // Seems Stable - repository TestMain procedure (unstable)
- add secondary indexes
- Docker compose instructions

- GetCount seems to fail on empty data...first start @TODO
``` // @TODO: FIX IT
curl -X GET -H "Authorization: Bearer $JWT" http://localhost:9090/api/v1/logs/count   
{}
```

### Run Application from docker
Build docker image:
```
docker build -t log-api .
```

Create a bridge docker network:
```
docker network create immudb-net
```
Immudb on docker bridged network:
```
docker run -it -d --net immudb-net -p 3322:3322 -p 9497:9497 --name immudb codenotary/immudb:latest
```
Run Log-API as:
```
docker run -it -d --net immudb-net -p 9000:9000 -p 9090:9090 -e immudb-host=immudb --name log_api log-api:latest ./app/api server 
```

Test it executing gRPC client or curl from local:
```
go run main.go client --token=$JWT foo_bar_key_1 auth-app "fake log content"          
2022/08/02 14:39:28 client called, arguments [foo_bar_key_1 auth-app fake log content] 
2022/08/02 14:39:28 Created Log Line key foo_bar_key_1_1659443968106459720
2022/08/02 14:39:28 History Key foo_bar_key_1_1659443968106459720 Revision [tx:3 value:"fake log content" revision:1]

```

Ensure JWT token is exported (use Login flow to get proper auth token)
```
curl -X GET -H "Authorization: Bearer $JWT" http://localhost:9090/api/v1/logs/count   
{"total":"1"}                                                                    
```

