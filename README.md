# go-structconf

[![license](https://img.shields.io/github/license/mashape/apistatus.svg?maxAge=2592000)](https://github.com/anexia-it/go-structconf/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/anexia-it/go-structconf?status.svg)](https://godoc.org/github.com/anexia-it/go-structconf)
![Test Status](https://github.com/anexia-it/go-structconf/actions/workflows/unit.yml/badge.svg)
![Lint Status](https://github.com/anexia-it/go-structconf/actions/workflows/code.yml/badge.svg)
[![codecov](https://codecov.io/gh/anexia-it/go-structconf/branch/master/graph/badge.svg)](https://codecov.io/gh/anexia-it/go-structconf)
[![Go Report Card](https://goreportcard.com/badge/github.com/anexia-it/go-structconf)](https://goreportcard.com/report/github.com/anexia-it/go-structconf)

go-structconf is a Go library which provides a simplified interface for
storing configuration data in structs.

It provides support for reading and writing configuration files from and to
storages as well as simplified usage of different data encodings.

Furthermore, it allows default values on config structs to be initialized
from structs of the same type.

## Install

With a [correctly configured](https://golang.org/doc/install#testing) Go toolchain:

```sh
go get -u github.com/anexia-it/go-structconf
```
