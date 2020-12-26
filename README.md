# go-cache
[![GoConcurrency](https://img.shields.io/badge/go-concurrency-blue)](https://tour.golang.org/concurrency/1)
[![GoDoc](https://godoc.org/github.com/sirupsen/logrus?status.svg)](https://golang.org/doc/)
[![Ask Me Anything !](https://img.shields.io/badge/Ask%20me-anything-1abc9c.svg)](https://github.com/parvez0)

A simple in memory cache store designed to simulate redis

### How to run
```shell
export NODE_ROLE="master"
go run main.go
```

### How to use

The server provides four operations get, set, lit and delete you can 
access the api's using following endpoint

```js
POST /set
{
    "key1": 1,
    "key2": "2",
    "key3": ["1", 2]
}
```
#### Response
```json
{
    "Inserted": 3,
    "Modified": 0,
    "Deleted": 0
}
```

```js
GET /get?key="key1"
```
#### Response
```json
1
```

```js
GET /list
```
#### Response
```json
{
  "key1": 1,
  "key2": "2",
  "key3": ["1", 2]
}
```
```js
Delete /delete?key="key3"
```
#### Response
```json
{
  "Inserted": 0,
  "Modified": 0,
  "Deleted": 1
}
```

### Whats next?

You can start a read only slave by the following command 
```shell
export NODE_ROLE="slave"
go run main.go
```
