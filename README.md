# coredns-docker

## Name

*docker* - enables serving zone data based on docker container names

## Description

The *docker* plugin is useful for development environments so you can access
containers by their names.

## Syntax
The configuration bellow will match containers by their names without using any
zone.
```
. {
  docker
  forward . 8.8.8.8
}
```
For instance, if you have a container named `my-nginx`, it will return
something like this:
```
my-nginx.   50      IN      A       172.25.0.2
```

In case you want `my-nginx` under a particular zone, you could use:
```
localdomain {
  docker
}
```
Then you would get:
```
my-nginx.localdomain   50      IN      A       172.25.0.2
```
