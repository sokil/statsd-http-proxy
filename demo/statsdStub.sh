#!/bin/bash

# This server emulates StatsD application.
# It gets messages from StatsD HTTP Proxy by UDP  and prints them to stdout

nc -kluv localhost 8125