package types

const CertificateType = "certificate"

type CertificateRequest struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Parameters struct {
		CommonName       string   `json:"common_name"`
		IsCa             bool     `json:"is_ca, omitempty"`
		Ca               string   `json:"ca, omitempty"`
		AlternativeNames []string `json:"alternative_names, omitempty"`
	}
}

func (r CertificateRequest) IsRootCaRequest() bool {
	return r.Type == CertificateType &&
		r.Parameters.IsCa &&
		r.Parameters.CommonName != "" &&
		r.Parameters.Ca == "" &&
		len(r.Parameters.AlternativeNames) == 0
}

func (r CertificateRequest) IsIntermediateCaRequest() bool {
	return r.Type == CertificateType &&
		r.Parameters.IsCa &&
		r.Parameters.CommonName != "" &&
		r.Parameters.Ca != "" &&
		len(r.Parameters.AlternativeNames) == 0
}

func (r CertificateRequest) IsRegularCertificateRequest() bool {
	return r.Type == CertificateType &&
		!r.Parameters.IsCa &&
		r.Parameters.Ca != "" &&
		r.Parameters.CommonName != ""
}

func (r *CertificateRequest) Generate() error {
	return nil
}

func (r *CertificateRequest) CredentialType() string {
	return r.Type
}

func (r *CertificateRequest) Validate() bool {
	return r.IsRootCaRequest() || r.IsIntermediateCaRequest() || r.IsRegularCertificateRequest()
}

type CertificatePostResponse struct {
}
