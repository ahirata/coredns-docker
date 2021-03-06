# coredns-docker
[![Build Status](https://travis-ci.org/ahirata/coredns-docker.svg?branch=master)](https://travis-ci.org/ahirata/coredns-docker)
[![Code Coverage](https://codecov.io/gh/ahirata/coredns-docker/branch/master/graph/badge.svg)](https://codecov.io/gh/ahirata/coredns-docker)
[![Go Report Card](https://goreportcard.com/badge/github.com/ahirata/coredns-docker)](https://goreportcard.com/report/ahirata/coredns-docker)

## Name

*docker* - enables serving zone data based on docker container names

## Description

The *docker* plugin is useful for development environments so you can access
containers by their names. This plugin is not intended for production usage.

## Syntax

The plugin is activated by using its name without any additional parameters
```
. {
  docker
}
```

## Example

The configuration bellow will match containers by their names without using any
zone:
```
. {
  docker
  forward . 8.8.8.8
}
```
For instance, if you have a container named `my-nginx`, it will return
something like this:
```
my-nginx.   0      IN      A       172.25.0.2
```

In case you want `my-nginx` under a particular zone, you could use:
```
localdomain {
  docker
}
```
Then you would get:
```
my-nginx.localdomain.   0      IN      A       172.25.0.2
```
