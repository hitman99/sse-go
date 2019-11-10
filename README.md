# Ultra Fast Server in Go for Server-Sent Events
This project uses Go modules, so in order to build it you will need to have  Go v1.11+

## Building

- `go build -o sse-server`

## Usage

 - build it first, see [Building](#building)
 - run it locally like this: `./sse-server`
 
More usage info (also available via `sse-server --help`)

```
Usage:
  sse-server [flags]

Flags:
  -h, --help               help for sse-server
      --host string        web server host (default "0.0.0.0")
      --port string        port for incoming connections (default "8080")
      --timeout duration   client timeout duration (default 30s)

```

## Testing
In order to run tests, just execute: `go test -cover ./...`, you should see similar output:

```
$ go test -cover ./...
?       github.com/hitman99/sse-go     [no test files]
?       github.com/hitman99/sse-go/cmd [no test files]
ok      github.com/hitman99/sse-go/internal/broker     0.786s  coverage: 94.3% of statements
?       github.com/hitman99/sse-go/internal/event      [no test files]
ok      github.com/hitman99/sse-go/internal/subscriber 2.476s  coverage: 88.6% of statements
?       github.com/hitman99/sse-go/internal/subscriber/mock    [no test files]
ok      github.com/hitman99/sse-go/internal/topic      0.861s  coverage: 100.0% of statements
```
