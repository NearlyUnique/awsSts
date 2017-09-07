package cmd

import "encoding/base64"

type (
	// Saml response body
	Saml string
)

// AsXML returns the decoded xml content
func (s Saml) AsXML() ([]byte, error) {
	xml, err := base64.StdEncoding.DecodeString(string(s))
	return xml, err
}

// AsAssertion returns the assertion blob to send to AWS AssumeRole
func (s Saml) AsAssertion() string {
	return string(s)
}
