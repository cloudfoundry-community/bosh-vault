# IDs
The config-server api requires all secrets to have a unique ID that points to a specific version of a given secret. Vault's [KV2 backend](https://www.vaultproject.io/docs/secrets/kv/kv-v2.html) implements versioned 
secrets already, however there is no "publicly" accessible concept of a UUID for a secret in Vault. Instead, secret versions
are requested at a given Vault path, for example: `vault kv get -version 3 secret/some_path/some_secret`. This is in contrast
to an implementation like Credhub which relies on a relational data store. 

There were a few ways to deal with this: 
  - Store multiple copies of the secret at different paths in Vault
  - Maintain some kind of map between UUIDs and Vault paths (either in Vault or elsewhere)
  - Take liberties with the length and form of an id

The first option is obviously problematic for several reasons. The key ones being: difficulty managing access control rules (breaks the security model), and the
duplicated effort required to accomplish something (secret versioning) that Vault is already doing for us. Likewise we felt maintaining some kind of map of UUIDs
to paths was not a good use of network traffic and prone to error. We opted for the third path and decided to have our IDs 
be base64 representations of credential metadata (name and version). Our understanding is that these ids are only used by 
the bosh director anyway, in which case their length is largely irrelevant. 


```
{
  name: "/Director/nginx/some_password",
  version: 1
}
```

would encode to:

```
eyJuYW1lIjoiL0RpcmVjdG9yL25naW54L3NvbWVfcGFzc3dvcmQiLCJ2ZXJzaW9uIjoxfQ==
```

A long an incomprehensible id to be sure, but it is guaranteed to be:
  - unique and in a 1-to-1 relationship with a given version of a secret
  - secretless (paths in Vault are not considered secret)
  - debuggable (base64 decode the ID to see what it was looking for)
