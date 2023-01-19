# Myrddin

[![Release](https://img.shields.io/github/release/taubyte/myrddin.svg)](https://github.com/taubyte/myrddin/releases)
[![License](https://img.shields.io/github/license/taubyte/myrddin)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/taubyte/myrddin)](https://goreportcard.com/report/taubyte/myrddin)
[![GoDoc](https://godoc.org/github.com/taubyte/myrddin?status.svg)](https://pkg.go.dev/github.com/taubyte/myrddin)
[![Discord](https://img.shields.io/discord/973677117722202152?color=%235865f2&label=discord)](https://tau.link/discord)

A powerful yaml template engine for golang.

## Usage
Import
```go
import "github.com/taubyte/myrddin"
```

Define a struct that will receive the data. Example:
```go
type MyConfig struct {
	Hostname string `yaml:"host"`
	Shell    string `yaml:"shell"`
	Name     string `yaml:"name"`
	Networks []struct {
		Name      string `yaml:"name"`
		AddrRange struct {
			Start string `yaml:"start"`
			End   string `yaml:"end"`
		} `yaml:"addr_range"`
	} `yaml:"networks"`
}
```

Create a myrddin instance and pass the recipient struct as an argument
```go
config := &MyConfig{}
m, err := myrddin.New(config)
```

You can aslo pass custom data and function as options
```go
m, err := myrddin.New(
    config,
    myrddin.Data(
        "ver",
        []string{"v0.0.1"},
    ),
    myrddin.Function(
        "unixTime",
        func() { return time.Now().Unix() },
    ),
)
```

Then, load the folder containing your files
```go
err = m.Load("config")
```

Finally, parse:
```go
err = m.Parse()
```


## Example
```bash
cd example
go run .
```
Output should look like
```yaml
host: hostname
shell: /bin/bash
name: test
networks:
    - name: net0
      addr_range:
        start: 192.168.0.1
        end: 192.178.0.100
    - name: net1
      addr_range:
        start: 192.168.1.1
        end: 192.178.1.100
    - name: net2
      addr_range:
        start: 192.168.2.1
        end: 192.178.2.100
```

## Maintainers
 - Samy Fodil @samyfodil
 - Sam Stoltenberg @skelouse
 - Aron Jalbuena @arontaubyte
