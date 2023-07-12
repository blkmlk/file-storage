## Distributed File Storage

### Description

A solution for the task to create an S3 like storage

### How to use?

1. Run the server
2. Send a request to /api/v1/upload to get an upload link
3. Send the file to the given upload link
4. Download the file from /api/v1/download/:file-name

### How to run tests?
```shell
make start
make test
make stop
```
