path "config-server/data/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "config-server/metadata/*" {
  capabilities = ["read"]
}

path "kv1/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}