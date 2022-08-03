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

## Design process

### Assumptions
- key/value mode does not offer Count command yet
- SQL mode does not offer detailed histories as required

### ImmuDB repository Design
Fist step has been design the immudb persistence layer and destination structures, requirements seems to point to the key/value implementation, as it solves Single & Batch LogLine insertions. (add & Sadd)
All LogLines History is achieved getting all created logLines using Scan command with an empty prefix.

Last N executed transactions can be resolved by getting ImmuDB current state, getting last Transaction ID, and then TxScan usage allows to get last N transactions. 

The corner case arrives at "Number of stored logs", my first understanding on that was using Count command, but, It seems not ready yet (I tested last master version from immudb pointing to v1.3.2-0.20220726101823-fb0428237486 version, getting same UnImplemented method).

As this is a requirement I found a compromise solution using schema preconditions, the idea is to use a unique ImmuDB key to store LogLines total number, use transactions on add & batch addition to control transactional behaviour.

The workflow goes like this:
- LogLine addition establish the precondition that logLine Key must not exist 
- In the transactional scope we get the current keySize value, I establish a secondary Precondition were the counter must be not changed concurrently
  - This precondition will fail on concurrent data insertion, right now is a trade-off solution, it needs more work to do to implement a full solution
  - A production ready solution will use a Compare and Swap approach (Optimistic concurrency), and so, a retry loop is required on precondition fail
    - planned solution with retry loop will be eventual consistent, under high concurrency this can be an issue, and potentially it will reduce the overall system performance.

#### Key Composition:
A log is defined by:
- Source
- Bucket
- Value
- time

Internally log lines are stored using key composition based on Source and T timestamp on nanoseconds, (avoid key overlapping is the overall goal here)
Key composition enables Scan usage to access prefixed keys, as example, get all logLines from a Source, and we can even narrow down in the Timestamp scope too
```
Example: 
  Source: foo_bar_key
  Timestamp: 1659437280251549814
Key: foo_bar_key_1659437280251549814
```

### Log Bucket Support
The same concatenation in key way can be used with Log Bucket names,  and so using Scan will enable us to filter all log lines by bucket, this will result in this key composition:
```
bucket: payment
key: payment_user_log_1659437280251549814
```
Scan probably has a penalty on performance as it will use iteration to get the result keys, reading on ImmuDB docs points to secondary indexes trough Sorted Sets, which seems a more effective solution to achieve Support for log buckets.
On this implementation: 
- A Sorted Set is defined by bucket
- Each LogLine key is added to the destination sorted set

Key addition in a sorted set will be idempotent, and so multiple additions wouldn't change zset composition. Thinking in transactional insertions Zsets are not supported with Sadd command, ExecAll it does, but it does not support Preconditions. In any case, as Zset idempotency is ensured, we can handle zset addition on Sadd Tranaction success, so that log line addition will remain using add & sadd.

On Zset scenario a better trade off is achieved (IMHO), using Zset we can separate logs from different applications using different buckets and Source and time can still be handled by key composition trough Scan command
Zset bucket inclusion has been added at the end of the development cycle, needs more test coverage but it works, we include log Line keys by prefix in the destination sorted set, and so we can get separated log lines by bucket.

### API implementation
The whole application has been designed as an API centric application, focused on gRPC proto definition and heavy usage of protoc compiler plugins.

Proposed solution uses:
- gRPC server 
- http REST server trough gRPC Gateway extension
- auth scheme implemented with JWT tokens as gRPC unary interceptor (stream interceptor is not implemented as no streams are defined yet)
  - auth API as been added to get jwt credentials over a faked User Repository (any user & password will succeed)
  - all API endpoints are restricted with the Login exception
  - gRPC and http client need to handle manually JWT token inclusion
    - client side interceptors seems the way to go to achieve full generation & renovation in a transparent manner

## Development flow and make it run
Info about the followed development path can be found in: ./doc/development.md

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

### Run test Suite
```
  go test --race ./...
```
**note: TestItInsertsMultipleLogLinesInBatch is Skipped until solve GetTxs() dirty key composition**

## Improvements
- Concurrent execution on Add & AddBatch is not supported
  - Precondition PreconditionKeyNotModifiedAfterTX will fail on concurrent writers
  - Compare and Swap (Optimistic Concurrency) will handle this fail using retry loop
  - Another way would be to linealize persistence request and handle them one by one (this will limit at concurrency to 1)
- errorGroups with context to handle grpc and http servers graceful shutdown
  - it will be needed to add end to end tests
- grpc client side auth interceptor
  - it can handle jwt credentials initialization and renewal from the client side in a transparent manner.

## TODO
- detected dirty Keys from tx.Entries (txs.GetTxs()), needs more investigation
  - full project consolidation after Zset secondary index addition
- add secondary indexes test coverage
  - feature added really quick, needs repository testing addition
- graceful gRPC & http server shutdown
  - use errorGroup with context on both transport server
  - handle application shutdown in a controlled manner 
  - use SigTerm and SigKill signals

- Increase test coverage
  - add test on relevant methods on the gRPC service layer (log) 
  - e2e test

- clean pending @TODOs
