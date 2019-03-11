# Contributing

Pull requests, issues, questions, etc. are welcome from all!

## Developer Workflow
This project is written in Go and the below workflow assumes you have a functional Go development environment setup that supports
the use of Go modules.

 1. Clone this repo and `cd` into it.
 1. Run `make` to see the available workflow commands.
 1. Run `make test` to run tests locally.

## Testing with a preexisting Bosh director or no Bosh director
 1. Run `make build`
 1. Write a custom configuration file (JSON or YAML) and pass it to the built binary via the `-config` flag
 
## Testing with a local Bosh director
To simplify the developer workflow this project relies on a tool called `blite` that makes it easy to get a bosh-lite 
environment setup on your local machine. The makefile will download the latest `blite` script [from the Zipcar/blite repo](https://github.com/Zipcar/blite) 
but you are expected to already have the [prerequisites](https://github.com/Zipcar/blite#dependencies) installed. 

`blite` will give useful error messages if it notices one or more prerequisites are missing. The makefile specifies a 
few test deployment scripts that will ensure things are working as expected, but once the director is up you're able to 
do anything you'd otherwise be able to do with a bosh director. Resources used to configure the local binary and local
director are in the `local-dev` directory.

### Option 1: All-In-One Test Deploys
 1. Run `make bosh-lite` to setup a local bosh-lite director running UAA and configured to communicate with a local bosh-vault binary
 1. Run `make vault` (in a new terminal window) to setup a local Vault server using the preconfigured storage backend at `local-dev/vault/data`
 1. Run `make unseal` to unseal the local development Vault
 1. Run `make run` to start the config server
 1. Run `make test-deploy-nginx` to deploy NGINX that will serve a single page that is filled with plain text credentials to show they can all be generated. 
 
### Option 2: Just The Director
 1. Run `make bosh-lite` to setup a local bosh-lite director running UAA and configured to communicate with a local bosh-vault binary
 1. Run `make vault` (in a new terminal window) to setup a local Vault server using the preconfigured storage backend at `local-dev/vault/data`
 1. Run `make unseal` to unseal the local development Vault
 1. Run `make run` to start the config server
 1. Run `eval $(./bin/blite env-eval)` to seed your terminal's environment with the credentials of your local bosh director so you can use standard `bosh` commands
 
### Option 3: Blite/bosh-vault Power User
 1. Make sure a Vault server is running and configured to use KV2
 1. Make sure the certs you want to use are in `local-dev/certs` (the next step will generate default certs if they don't exist)
 1. Run the compiled binary using your desired configuration (or the default in `local-dev/config`)
 1. Run `blite create` passing in operator and vars files using the `BLITE_OPS_FILE_GLOB` and `BLITE_VARS_FILE_GLOB` environment variables, at a minimum you'll need what is captured in `local-dev/operators` and `local-dev/vars`. Alternatively just add operator/vars files directly to those directories using the same naming convention.
 1. Run `eval $(./bin/blite env-eval)` to seed your terminal's environment with the credentials of your local bosh director so you can use standard `bosh` commands.
 1. Use bosh to set a custom cloud config (or use `blite cloud-config` for default settings)
 1. Use as a normal bosh director you power user you!
 
## Local Vault
This repo contains the file storage backend for a development/testing Vault in `local-dev/vault/data`. There are Make
scripts that will utilize it (`make vault` and `make unseal`) or you can use it directly. Checkout the [local Vault README
file](https://github.com/Zipcar/bosh-vault/blob/master/local-dev/vault/README.md).
 
## Troubleshooting Dev Workflow Issues

#### Certificate Problems
Certificate problems are the most common cause of local dev frustration. Bosh requires config servers use TLS; to try
and make local development as simple as possible there is no port forwarding or DNS level magic happening on the bosh-lite
director. Instead the binary is bound on `0.0.0.0`, certs are generated for the hosts primary LAN IP (obtained using routes table),
and the director is configured to talk to the host machine over it's primary LAN IP. This works well until your IP changes;
say from a VPN connection, new wifi network, etc. For the moment the best way to deal with this issue is to run `make destroy`
and then start over with one of the local Bosh director options specified above. It should only take a few minutes to spin up 
a new director with fresh certs. 

For the brave, it is possible to manually fix cert mismatch problems by deleting in `local-dev/certs/local-dev.*`, running 
`make local-certs` to generate new ones, replacing `/var/vcap/jobs/director/config/config_server_ca.cert` and running 
`monit restart all` on the director then restarting bosh-vault with `make run` but it's probably a better idea to just 
destroy and recreate, it's fast and less error prone. Note that if you're reusing a shell session you may need to re-run
`eval $(./bin/blite env-eval)` to ensure your `bosh` command can communicate with the new director.

#### Networking Problems
`blite` provides networking helpers (like [`blite route-add`](https://github.com/Zipcar/blite#route-add)) that ensure your host machine can communicate with the things you deploy with a bosh-lite director.
Run `./bin/blite networking` to see what your current bosh-lite networking configuration is (default is bosh-lite default 
from the bosh-deployment repo). As mentioned in the [blite documentation](https://github.com/Zipcar/blite#avoiding-network-issues) your local network, VPN, etc
might conflict. If that's the case you can override the environment variables listed by the `networking` command and redeploy. 
If your networking issue is constant, say due to VPN routes conflicting, you should override those environment variables 
permanently in something like `~/.bashrc` or `~/.bash_profile` to ensure `blite` will always work properly.

# Resources
  - [Config server api documentation](https://github.com/cloudfoundry/config-server/blob/master/docs/api.md)
  - [Credhub Implementation Docs](http://credhub-api.cfapps.io/version/2.1/)