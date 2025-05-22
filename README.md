[docimg]:https://godoc.org/github.com/ardnew/mung?status.svg
[docurl]:https://godoc.org/github.com/ardnew/mung
[repimg]:https://goreportcard.com/badge/github.com/ardnew/mung
[repurl]:https://goreportcard.com/report/github.com/ardnew/mung
[covimg]:https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/ardnew/8642d8c0268d7a0f7e436e66dbdbbf88/raw/badge-mung-coverage.json
[covurl]:https://gist.githubusercontent.com/ardnew/8642d8c0268d7a0f7e436e66dbdbbf88/raw/badge-mung-coverage.json


[release packages]:https://github.com/ardnew/mung/releases
[download]:https://github.com/ardnew/mung/releases/latest

# mung
#### Manipulate PATH-like environment variables

[![GoDoc][docimg]][docurl] [![Go Report Card][repimg]][repurl] [![Test Coverage][covimg]][covurl]

The `mung` package contains a reusable [Go module](.) and [command-line utility](cmd/mung) ([download]).
It is fully documented (with examples) and includes tests with 100% coverage.

The comand-line utility, also named `mung`, is a simple wrapper around the module.
Like most native Go executables, the [release packages] are all statically-linked and have no dependencies.
It is convenient for shell script usage.

## Usage

See [Go doc](https://godoc.org/github.com/ardnew/mung) for all documentation and examples.

## Installation

To use the module in your own Go packages:

```sh
go get -v github.com/ardnew/mung
```

To install the command-line utility:

```sh
go install -v github.com/ardnew/mung/cmd/mung@latest
```

## Command-line Examples (GNU bash)

```sh
set -x

# prefix "bar" a given string "foo"
export X=$( mung -p bar foo )
++ mung -p bar foo
+ export X=bar:foo
+ X=bar:foo

# move "foo" to prefix the env var named X
export X=$( mung -p foo -n X )
++ mung -p foo -n X
+ export X=foo:bar
+ X=foo:bar

# suffix "baz" to X
export X=$( mung -s baz -n X )
++ mung -s baz -n X
+ export X=foo:bar:baz
+ X=foo:bar:baz

# multiple operations to show ordering and replacement
export X=$( mung -p two:three -p one -s x:y -s z -r baz -n X )
++ mung -p two:three -p one -s x:y -s z -r baz -n X
+ export X=one:two:three:foo:bar:x:y:z
+ X=one:two:three:foo:bar:x:y:z

# another complex example
export X=$( mung -s front:end -r next:bar -p front:next -n X )
++ mung -s front:end -r next:bar -p front:next -n X
+ export X=front:one:two:three:foo:x:y:z:end
+ X=front:one:two:three:foo:x:y:z:end
```

# Packaging releases

A [`Makefile`](Makefile) is provided that will generate versioned release packages.

Define `VERSION` and use the `dist` target to build for all platforms:

```sh
make dist VERSION=X.Y.Z
```

To build/package for a specific platform, call `make [<target>-]${GOOS}-${GOARCH}`. For example with a single platform `linux-amd64`:

```sh
make     clean-linux-amd64
make           linux-amd64 VERSION=X.Y.Z
make      dist-linux-amd64 VERSION=X.Y.Z
make distclean-linux-amd64 VERSION=X.Y.Z
