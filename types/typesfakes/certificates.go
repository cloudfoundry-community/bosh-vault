package typesfakes

const RootCaRequestBody = `
{
  "name": "my_cert",
  "type": "certificate",
  "parameters": {
    "is_ca": true,
    "common_name": "bosh.io"
  }
}
`

const IntermediateCaRequestBody = `
{
  "name": "my_cert",
  "type": "certificate",
  "parameters": {
    "is_ca": true,
    "ca": "my_ca",
    "common_name": "bosh.io"
  }
}
`

const RegularCertRequestBody = `
{
  "name": "my_cert",
  "type": "certificate",
  "parameters": {
    "ca": "my_ca",
    "common_name": "bosh.io",
    "alternative_names": ["bosh.io", "blah.bosh.io"]
  }
}
`
