# Bosh Vault
This repo is an implementation of the [config server API](ttps://github.com/cloudfoundry/config-server/blob/master/docs/api.md) using Vault as a backend. It is geared towards teams that already 
have working secrets management with Vault and want to leverage that for their Bosh infrastructure in lieu of Credhub.

There is a functional [Bosh release for this project](https://github.com/Zipcar/bosh-vault-release/releases) but it can also be run as a standalone
binary outside of Bosh.

[![CircleCI](https://circleci.com/gh/Zipcar/bosh-vault/tree/master.svg?style=svg)](https://circleci.com/gh/Zipcar/bosh-vault/tree/master)

# Configuration
The bosh-vault binary can be configured using a configuration file or environment variables. In the case where both are 
provided environment variables will override configuration file settings.

A configuration file can be passed using the flag: `-config` and passing a path to a JSON or YAML file of the form:

```
api:
  address: 0.0.0.0:1337 (Binding for the config-server API)
  draintimeout: 30 (How many seconds the config server should drain connections when shutting down)
log:
  level: ERROR (ERROR | INFO | DEBUG)
vault:
  address: NO_DEFAULT (Address of a Vault server with KV2 mount available for config-server to use)
  token: NO_DEFAULT (Token that allows data and metadata access on config-server's KV2 mount; periodic token suggested)
  timeout: 30 (How many seconds to wait when contacting Vault before timing out)
  mount: secret (The name of the KV2 mount in Vault)
  ca: NO_DEFAULT (Path to the CA to trust when connecting to Vault)
  skipverify: false (Whether or not to skip verifying TLS trust)
  renewinterval: 3600 (How many seconds to wait before renewing the vault token)
tls:
  cert: NO_DEFAULT (Path to the cert used to secure the config server api)
  key: NO_DEFUAULT (Path to the key used to secure the config server api)
uaa:
  enabled: true (Whether or not the config server should require and verify UAA JWT tokens)
  address: NO_DEFAULT (The address of the UAA server to communicate with)
  timeout: 10 (How many seconds to wait before timing out connections to UAA)
  ca: NO_DEFAULT (Path to the CA to trust when connecting to UAA)
  skipverify: false (Whether or not to skip verifying TLS trust)
  audienceclaim: config_server (Expected audience claim on a given JWT)
  keyrefreshinterval: 86400 (How many seconds to wait before fetching updated public key info from UAA) 
```

These variables can also be passed on the environment by prefixing them with `BV` and using underscores. For example to 
pass the uaa address: `BV_UAA_ADDRESS`

## Configuring Vault Storage
Bosh-vault requires a Vault server with a [KV2 mount](https://www.vaultproject.io/docs/secrets/kv/kv-v2.html) available.
```
vault secrets enable -version=2 -path=config-server kv
```

In order for bosh-vault to work with an existing Vault server it needs a token. That token should be attached to a 
policy that looks something like this (assuming your KV2 mount was `config-server`):

```
path "config-server/data/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "config-server/metadata/*" {
  capabilities = ["read"]
}
```

```
vault policy write config-server config-server.hcl
```

You'll also need to generate a token for the config server to use that is tied to this policy, we recommend using a 
periodic token as bosh-vault can be configured to automatically renew its token.

```
vault token create -format=json -period=168h -policy=config-server -display-name=bosh-vault-config-server
```

## Configuring UAA Auth

By default bosh-vault expects to receive a JWT token for authentication that has an audience claim of `config_server`.
Here's an example operator entry to configure a client.

```
- type: replace
  path: /instance_groups/name=bosh/jobs/name=uaa/properties/uaa/clients/director_config_server?
  value:
    override: true
    authorized-grant-types: client_credentials
    scope: ""
    authorities: config_server.admin, uaa.resource
    secret: some-good-password-1-2-3-4-5-6
```

# Redirect Pull Through Cache
This implementation of config server supports a feature that is not in the API spec or CredHub implementation: redirects.
Redirects are meant to provide a means to operationalize some of Vaults most powerful features via config-server endpoints.

All redirects copy the redirect value into the default Vault every time they are requested. This "last known" value 
will be returned only if the configured redirect vault is unhealthy (sealed, down, etc).

Redirects are only evaluated for GET requests. PUT, POST, and DELETE requests never apply redirect logic and operate against
the "local" Vault. Any changes applied "locally" to secrets with a configured redirect will be overwritten on the next 
successful redirected GET request.

Configure redirects with the following syntax:

```
redirects:
  - type: upstream (v1 | dynamic | upstream)
    vault:
      address: NO_DEFAULT
      token: NO_DEFAULT
      timeout: 30 (How many seconds we should wait when contacting Vault before timing out)
      mount: secret (The name of the KV2 mount in Vault, not used for v1 or dynamic redirects)
      ca: NO_DEFAULT (Path to the CA to trust when connecting to Vault)
      skipverify: false (Whether or not to skip verifying TLS trust)
      renewinterval: 3600 (How many seconds we should wait before renewing our vault token)
    rules :
    - ref: /DIRECTOR_NAME/DEPLOYMENT_NAME/star_yourdomain_biz
      redirect: /global/certificate/star.yourdomain.biz
    - ref: /DIRECTOR_NAME/DEPLOYMENT_NAME/a_shared_credential
      redirect: /global/password/a_shared_credential
 
```

Note redirects is an array where multiple sources and types can be specified. Pattern matching and wild carding on refs 
is explicitly NOT supported. If two rules collide the first one will be followed.

## Redirect Types
Three types of redirects are supported:
  
### upstream

Upstream redirects are meant to provide a way to request credentials from different Vault servers or paths that are already
using KV2. Access can be controlled with a unique token and policy to accomplish things like "read-only" credentials. This
also allows for centralized management and auditing of certain credentials. Fetched values are cached in the default Vault
store at the expected ref and in the event the upstream can't be reached the local "last-known" value will be returned.

### v1

Migrating from KV1 to KV2 can be challenging depending on your existing Vault configuration and usage. This redirect is 
meant to fetch values from static locations in Vault that are still using KV1. This feature's caching behavior functions 
just like upstream redirects and is an ideal migration path to KV2 and versioned infrastructure secrets. Once a single 
deployment has pulled the values from KV1 the redirect can be removed and config server will operate as normal on the 
newly created KV2 entry at the expected path.

### dynamic

Vault supports lots of dynamic secret engines which can both generate and expire credentials for many external services. 
Dynamic redirects allow config server to take advantage of these. For example, let's say you configured a postgres DB 
role in Vault to generate expiring access to postgres at: `database/creds/my-app-role`

In your template you could then request: 
```
PGUSER: ((dynamic_postgres.username))
PGPASS: ((dynamic_postgres.password))
```

Using the following redirect rule:
```
redirects:
  - type: "dynamic"
    rules:
      - ref: "/BoshDirectorName/my_app_deployment/dynamic_postgres"
        redirect: "database/creds/my-app-role"
    vault: *vault
```

New credentials will be fetched on each deploy and Vault will expire the old ones according to TTLs managed by your Vault
team. Effectively solving credential rotation in cases where Bosh can get creds from one of Vault's supported secret engines. 

# Deployment Architecture
The bosh-vault config server implementation is meant to be run alongside Vault and proxy config server requests. It could 
also be located on the director as a job using the bosh-release but this has security implications as it would mean storing 
the token for the config server on the director itself. It could also be deployed as its own Bosh deployment or a stand alone 
binary outside of Bosh but this can make management a little more cumbersome.

![high level architecture diagram](docs/diagrams/high-level-architecture.jpg)

Because bosh-vault is stateless and already dependent on a healthy Vault server, running it as a sidecar to Vault makes
management relatively straightforward. In cases where you're using HA Vault deployed behind a load balancer with DNS and
health checks, ensure that each Vault instance is also running a sidecar bosh-vault and send director requests through 
the load balancer; Then your config-server setup will be just as resilient as your Vault deployment.
