# ubirch-go-c8y-client [WIP]

## Instalation

`go get github.com/ubirch/ubirch-go-c8y-client/c8y`

## Usage

`import "github.com/ubirch/ubirch-go-c8y-client/c8y"`

## Run example
- Create `config.json` in `main` directory:
```json
{
  "uuid": "<UUID>",
  "bootstrap": "<cumulocity password>",
  "tenant": "<tenantID>"
}
```

- Register UUID in the Device Management application in Cumulocity

- Build and run the example
```
$ cd $HOME/go/src/github.com/ubirch/ubirch-go-c8y-client/main
$ go build
$ ./main
```

- Accept the connection from the device in the Device Registration page in Cumulocity