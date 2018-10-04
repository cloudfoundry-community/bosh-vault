# Vault CloudFoundry Config Server
This repo is an attempt to make an implementation of the config server API using Vault as a backend. It is a work in progress and should not be used.

[![CircleCI](https://circleci.com/gh/Zipcar/vault-cfcs/tree/master.svg?style=svg)](https://circleci.com/gh/Zipcar/vault-cfcs/tree/master)

# Resources
  - [Config server api documentation](https://github.com/cloudfoundry/config-server/blob/master/docs/api.md)
  - [Bosh config server integration with credhub](https://github.com/cloudfoundry-incubator/credhub/blob/master/docs/bosh-config-server.md)
  - [BUCC config server operator file](https://github.com/starkandwayne/bucc/blob/d477e927c79014b86a8694f3d724f260ae9f2fff/src/bosh-deployment/misc/config-server.yml)
 
# Architecture
The Vault Cloudfoundry Config Server is meant to be run alongside Vault and proxy config server requests.

![high level architecture diagram](docs/diagrams/high-level-architecture.jpg)

# Contributing
 1. Clone this repo into `$GOPATH/src/github.com/zipcar/vault-cfcs`
 1. Run `make` to see the available workflow commands.
 1. Run `make test` and `make run` to get things running locally.