[docimg]:https://godoc.org/github.com/ardnew/mung?status.svg
[docurl]:https://godoc.org/github.com/ardnew/mung
[repimg]:https://goreportcard.com/badge/github.com/ardnew/mung
[repurl]:https://goreportcard.com/report/github.com/ardnew/mung
[covimg]:https://codecov.io/gh/ardnew/mung/branch/main/graph/badge.svg
[covurl]:https://codecov.io/gh/ardnew/mung

# mung
#### Manipulate PATH-like environment variables

[![GoDoc][docimg]][docurl] [![Go Report Card][repimg]][repurl] [![codecov][covimg]][covurl]

The `mung` package contains a reusable module with types, methods, and so on.
It is fully documented and includes test cases with 100% coverage.

There is a companion command-line utility, also named `mung`, which is a simple
wrapper around the module. It is convenient for shell script usage.

## Usage

See [Go doc](https://godoc.org/github.com/ardnew/mung) for now.

## Installation

To use the module in your own Go packages:

```sh
go get -v github.com/ardnew/mung
```

To install the command-line utility:

```sh
go install -v github.com/ardnew/mung/cmd/mung@latest
```

Future releases will include binary distribution packages.
