package storefakes

import (
	"encoding/json"
	"github.com/cloudfoundry-community/bosh-vault/store"
)

var ValidSecretMetadata = store.VersionedSecretMetaData{
	Name:    "/DatDirector/DatDeployment/DatVar",
	Version: json.Number("1"),
}

var ValidSecretMetadataId = "eyJuYW1lIjoiL0RhdERpcmVjdG9yL0RhdERlcGxveW1lbnQvRGF0VmFyIiwidmVyc2lvbiI6MX0="
