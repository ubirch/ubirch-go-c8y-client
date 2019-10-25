# [WIP] ubirch-go-c8y-client

## Instalation

`go get github.com/ubirch/ubirch-go-c8y-client/c8y`

## Usage

`import "github.com/ubirch/ubirch-go-c8y-client/c8y"`

## Run example
> This example can be used to bootstrap and retrieve device credentials from Cumulocity.
- Create `config.json` in `main` directory:
```json
{
  "uuid": "<UUID>",
  "tenant": "<cumulocity tenant ID>",
  "bootstrap": "<cumulocity password>"
}
```

- Register UUID in the Device Management application in Cumulocity

- Build and run the example
```
$ cd main
$ go build
$ ./main
```

- Accept the connection from the device in the Device Registration page in Cumulocity

- The device credentials will be saved to a file in `main` directory: `<<UUID>>.json`