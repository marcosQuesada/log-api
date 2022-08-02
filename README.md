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

## Assumptions


## Improvements
- next steps....
  - Precondition fail on concurrent writers needs a retry loop
  - grpc client side auth interceptor
- errorGroups with context to handle grpc and http servers


## TODO
- e2e test
  // Seems Stable - repository TestMain procedure (unstable)
- add secondary indexes