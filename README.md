# statsd-rest-server
HTTP Server with REST interface to StatsD

[![Go Report Card](https://goreportcard.com/badge/github.com/sokil/statsd-rest-server?1)](https://goreportcard.com/report/github.com/sokil/statsd-rest-server)

This server is a HTTP proxy to UDP connection. Usefull for sending tracking to StatsD from frontend by AJAX. Authentication based on jwt token.

## Usefull resources
* [https://github.com/etsy/statsd](https://github.com/etsy/statsd) - StatsD sources
* [Docker image with StatsD, Graphite, Grafana 2 and a Kamon Dashboard](https://github.com/kamon-io/docker-grafana-graphite)

## Installation

```
got clone git@github.com:sokil/statsd-rest-server.git
go get
go build
```

## Useage

Server options:
```
statsd-rest-server --http-host=127.0.0.1 --http-pport=8080 --statsd-host=127.0.0.1 --statsd-port=8125 --jwt-secret=somesecret
```

## Authentication

Token must be encrypted with secret, specifier in passed in `jwt-secret` of server. Token sends to server in `X-JWT-Token` header. If server started without passing JWT sectet in option `jwt-secret` then requests to server accepred without authentication.

## Rest resources

### Heartbeat
```
GET /heartbeat
```
If server working, it responds with `OK`

### Count
```
POST /count/{key}
X-JWT-Token: {tokenString}
delta=1&sampleRate=1
```

| Parameter  | Description                          | Default value                      |
|------------|--------------------------------------|------------------------------------|
| delta      | Value of delta. Negative to decrease | Optional. Default 1                |
| sampleRate | Sample rate to skip metrics          | Optional. Default to 1: accept all |

### Gauge
```
POST /gauge/{key}
X-JWT-Token: {tokenString}
value=1
```

| Parameter  | Description                          | Default value                      |
|------------|--------------------------------------|------------------------------------|
| value      | Value                                | Optional. Default 1                |

### Timing
```
POST /timing/{key}
X-JWT-Token: {tokenString}
time=1234567&sampleRate=1
```

| Parameter  | Description                          | Default value                      |
|------------|--------------------------------------|------------------------------------|
| time       | Time in milloiseconds                | Required                           |
| sampleRate | Sample rate to skip metrics          | Optional. Default to 1: accept all |

### Set
```
POST /set/{key}
X-JWT-Token: {tokenString}
value=1
```

| Parameter  | Description                          | Default value                      |
|------------|--------------------------------------|------------------------------------|
| value      | Value                                | Optional. Default 1                |
