# StatsD HTTP Proxy

HTTP Server with REST interface to StatsD

[![Go Report Card](https://goreportcard.com/badge/github.com/sokil/statsd-http-proxy?2)](https://goreportcard.com/report/github.com/sokil/statsd-http-proxy)
[![Build Status](https://travis-ci.org/sokil/statsd-http-proxy.svg?branch=master)](https://travis-ci.org/sokil/statsd-http-proxy)
[![Code Climate](https://codeclimate.com/github/sokil/statsd-http-proxy/badges/gpa.svg?1)](https://codeclimate.com/github/sokil/statsd-http-proxy)

This server is a HTTP proxy to UDP connection.
Useful for sending metrics to StatsD from frontend by AJAX.
Authentication based on JWT token.

## Table of contents

* [Installation](#installation)
* [Requirements](#requirements)
* [Usage](#usage)
* [Authentication](#authentication)
* [Rest resources](#rest-resources)
  * [Heartbeat](#heartbeat)
  * [Count](#count)
  * [Gauge](#gauge)
  * [Timing](#timing)
  * [Set](#set)
* [Response](#response)
* [Testing](#testing)
* [Benchmark](#benchmark)
* [Useful resources](#useful-resources)


## Installation

```
got clone git@github.com:sokil/statsd-http-proxy.git
make build
```

Also available [Docker image](https://hub.docker.com/r/sokil/statsd-http-proxy/):

```
docker run -p 80:80 sokil/statsd-http-proxy:latest --verbose
```

## Requirements

* [sokil/go-statsd-client](https://github.com/sokil/go-statsd-client) - StatsD client library for Go
* [dgrijalva/jwt-go](https://github.com/dgrijalva/jwt-go) - JSON Web Tokens builder and parser
* [gorilla/mux](https://github.com/gorilla/mux) - URL router and dispatcher
	
## Usage

Run server:

```bash
statsd-http-proxy \
	--verbose \
	--http-host=127.0.0.1 \
	--http-port=8080 \
	--statsd-host=127.0.0.1 \
	--statsd-port=8125 \
	--jwt-secret=somesecret \
	--metric-prefix=prefix.subprefix
```

Print server version and exit:

```bash
statsd-http-proxy --version
```

Command line arguments:

| Parameter       | Description                          | Default value                                                                  |
|-----------------|--------------------------------------|--------------------------------------------------------------------------------|
| verbose         | Print debug info to stderr           | Optional. Default false                                                        |
| http-host       | Host of HTTP server                  | Optional. Default 127.0.0.1. To accept connections on any interface, set to "" |
| http-port       | Port of HTTP server                  | Optional. Default 80                                                           |
| statsd-host     | Host of StatsD instance              | Optional. Default 127.0.0.1                                                    |
| statsd-port     | Port of StatsD instance              | Optional. Default 8125                                                         |
| jwt-secret      | JWT token secret                     | Optional. If not set, server accepts all connections                           |
| metric-prefix   | Prefix, added to any metric name     | Optional. If not set, do not add prefix                                        |
| version         | Print version of server and exit     | Optional                                                                       |

## Authentication

Token must be encrypted with secret, specified in CLI parameter `jwt-secret` of server.
Token sends to server in `X-JWT-Token` header.
If JWT secret not configured in `jwt-secret` parameter on server start,
then requests to server accepted without authentication.

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
value=1&sampleRate=1
```

| Parameter  | Description                          | Default value                      |
|------------|--------------------------------------|------------------------------------|
| value      | Value. Negative to decrease          | Optional. Default 1                |
| sampleRate | Sample rate to skip metrics          | Optional. Default to 1: accept all |

### Gauge

Gauge is an arbitrary value. 
Only the last value during a flush interval is flushed to the backend. If the gauge is not updated at the next flush, it will send the previous value.
Gauge also may be set relatively to previously stored value. Is shift not set, then checked value. 
If value not sed, used default value equals 1.

Absolute value:
```
POST /gauge/{key}
X-JWT-Token: {tokenString}
value=1
```

Shift of previous value:
```
POST /gauge/{key}
X-JWT-Token: {tokenString}
shift=-1
```

| Parameter  | Description                                     | Default value                      |
|------------|-------------------------------------------------|------------------------------------|
| value      | Integer value                                   | Optional. Default 1                |
| shift      | Signed int, relative to previously stored value | Optional                           |

### Timing
```
POST /timing/{key}
X-JWT-Token: {tokenString}
time=1234567&sampleRate=1
```

| Parameter  | Description                                   | Default value                      |
|------------|-----------------------------------------------|------------------------------------|
| time       | Time in milliseconds                          | Required                           |
| sampleRate | Float sample rate to skip metrics from 0 to 1 | Optional. Default to 1: accept all |

### Set
```
POST /set/{key}
X-JWT-Token: {tokenString}
value=1
```

| Parameter  | Description                          | Default value                      |
|------------|--------------------------------------|------------------------------------|
| value      | Integer value                                | Optional. Default 1                |

## Response

Server sends `200 OK` if send success, even StatsD server is down.

Other HTTP status codes:

| CODE             | Description                             |
|------------------|-----------------------------------------|
| 400 Bad Request  | Invalid paraameters specified           |
| 401 Unauthorized | Token not sent                          |
| 403 Forbidden    | Token invalid/expired                   |
| 404 Not found    | Invalid url requested                   |
| 405 Wrong method | Request method not allowed for resource |

## Testing

It is useful for testing to start `netcat` UDP server,
listening for connections and watch incoming metrics.
To start server run:
```
nc -kluv localhost 8125
```

## Benchmark

Machine for benchmarking:

```
Intel(R) Core(TM) i5-2450M CPU @ 2.50GHz Dual Core / 8 GB RAM
```

Siege test:

```
$ GOMAXPROCS=2 ./bin/statsd-http-proxy --verbose --http-host=127.0.0.1 --http-port=8080 --statsd-host=127.0.0.1 --statsd-port=8125 --jwt-secret=somesecret

$ time siege -c 255 -r 255 -b -H 'X-JWT-Token:eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJzdGF0c2QtcmVzdC1zZXJ2ZXIiLCJpYXQiOjE1MDY5NzI1ODAsImV4cCI6MTg4NTY2Mzc4MCwiYXVkIjoiaHR0cHM6Ly9naXRodWIuY29tL3Nva2lsL3N0YXRzZC1yZXN0LXNlcnZlciIsInN1YiI6InNva2lsIn0.sOb0ccRBnN1u9IP2jhJrcNod14G5t-jMHNb_fsWov5c' "http://127.0.0.1:8080/count/a.b.c.d POST value=42"
  ** SIEGE 4.0.2
  ** Preparing 255 concurrent users for battle.
  The server is now under siege...
  Transactions:                  65025 hits
  Availability:                 100.00 %
  Elapsed time:                  19.64 secs
  Data transferred:               0.00 MB
  Response time:                  0.05 secs
  Transaction rate:            3310.85 trans/sec
  Throughput:                     0.00 MB/sec
  Concurrency:                  180.67
  Successful transactions:       65025
  Failed transactions:               0
  Longest transaction:            1.37
  Shortest transaction:           0.00


  real    0m19.694s
  user    0m6.068s
  sys     0m38.440s
```

## Useful resources
* [https://github.com/etsy/statsd](https://github.com/etsy/statsd) - StatsD sources
* [Official Docker image for Graphite](https://github.com/graphite-project/docker-graphite-statsd)
* [Docker image with StatsD, Graphite, Grafana 2 and a Kamon Dashboard](https://github.com/kamon-io/docker-grafana-graphite)
* [Online JWT generator](http://jwtbuilder.jamiekurtz.com/)
