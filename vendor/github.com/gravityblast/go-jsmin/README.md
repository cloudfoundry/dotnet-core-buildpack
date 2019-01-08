go-jsmin
========

Douglas Crockford's JSMin in Go. Package and command. #golang

## Usage

The command accepts a stream from the standard input and prints the compressed
version on the standard output.

```
$ go get github.com/web-assets/go-jsmin/cmd/gojsmin
$ gojsmin < original.js > compressed.js
```

## Development setup

```
$ mkdir -p $GOPATH/src/github.com/web-assets
$ git clone https://github.com/web-assets/go-jsmin $GOPATH/src/github.com/web-assets/go-jsmin && cd $_
$ make setup
$ make
```
