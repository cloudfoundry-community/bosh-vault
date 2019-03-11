# Local Dev Vault

This directory contains everything needed to run a Vault server in a known state for local development and experimentation. 
It is not meant to store real credentials or be reused in any way for any reason.

Because of the way the Vault file storage backend works this development Vault has its state semi-frozen using `.gitignore`.
New entries in this Vault cause new files to be created in `local-dev/vault/data/` which by default are ignored. This is
nice because they will not be committed to the repository, but they will persist on the local development machine even 
through restarts. If you want/need to clear the state of this Vault server back to it's default or "frozen" state run 
`make reset-vault` which will remove any new entries and reset any base configuration changes.

## Usage
The best way to use this Vault is with the make file: `make vault`
You will also have to unseal it: `make unseal`

However if you're a DIY'r you will have you best results by **CDing into this directory** and then running:

```
vault server -config config.hcl
```

## Access
You know dem bots are going to have a field day with this one.

```
vault operator unseal tjx5+szUOw96N6e3Ge2ss+YPFXJVoa2XkwC7h5ZJJfY=
vault login s.48t1wu9P4mLBgzvA1LOMJ7AV
```

When running the local bosh-vault binary via `make run` it will not use the root token above, instead it will generate a 
periodic token attached to the config-server policy and use that.

## Secret Stores
A KV2 mount called `config-server`
A KV1 mount called `kv1`

```
matt$ vault secrets list
Path              Type         Accessor              Description
----              ----         --------              -----------
config-server/    kv           kv_38f3a6ef           mount for config-server binary
cubbyhole/        cubbyhole    cubbyhole_0780464b    per-token private secret storage
identity/         identity     identity_f55b4409     identity store
kv1/              kv           kv_dd8dfdd3           key/value secret storage
sys/              system       system_430e496e       system endpoints used for control, policy and debugging

```

## Audit Devices
This test vault is setup to send audit data to stdout. This was configured with the following command:

```
vault audit enable file file_path=stdout
```