package types

type CertificatePostRequest struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Parameters struct {
		CommonName       string   `json:"common_name"`
		IsCa             bool     `json:"is_ca, omitempty"`
		Ca               string   `json:"ca, omitempty"`
		AlternativeNames []string `json:"alternative_names, omitempty"`
	}
}

func (r CertificatePostRequest) IsRootCaRequest() bool {
	return r.Parameters.IsCa &&
		r.Parameters.CommonName != "" &&
		r.Parameters.Ca == "" &&
		len(r.Parameters.AlternativeNames) == 0
}

func (r CertificatePostRequest) IsIntermediateCaRequest() bool {
	return r.Parameters.IsCa &&
		r.Parameters.CommonName != "" &&
		r.Parameters.Ca != "" &&
		len(r.Parameters.AlternativeNames) == 0
}

func (r CertificatePostRequest) IsRegularCertificateRequest() bool {
	return !r.Parameters.IsCa &&
		r.Parameters.Ca != "" &&
		r.Parameters.CommonName != ""
}

type CertificatePostResponse struct {
}
