package cmd

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/pkg/errors"
)

type (
	// SamlDocument contains the payload of the STS authentication
	SamlDocument struct {
		XMLName   xml.Name         `xml:"Response"`
		Assertion []AttributeValue `xml:"Assertion>AttributeStatement>Attribute"`
	}
	// AttributeValue contains the core information for role based assertion
	AttributeValue struct {
		Name  string   `xml:"Name,attr"`
		Value []string `xml:"AttributeValue"`
	}
	//Arn principal and role
	Arn struct {
		// show to user
		principal, role, alias string
		// store in config
		accessKeyID, secretAccessKey, sessionToken string
	}
)

//ExtractRoles from the saml single sign on response
func ExtractRoles(saml *Saml, cache *AccountAliasCache, durationSeconds int64) (arns []Arn, err error) {
	var xml []byte
	xml, err = saml.AsXML()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read SAML")
	}
	arns, err = extractRoleAttributes(xml)
	if err != nil {
		return arns, err
	}
	lookupAccountAliases(arns, saml.AsAssertion(), cache, durationSeconds)
	return arns, err
}

func extractRoleAttributes(raw []byte) ([]Arn, error) {
	saml := SamlDocument{}
	err := xml.Unmarshal(raw, &saml)
	dumpFile("saml-xml-", raw)
	if err != nil {
		return nil, errors.Wrap(err, "Reading SAML XML")
	}
	roles := []Arn{}
	for _, a := range saml.Assertion {
		if a.isRole() {
			roles = a.arns()
			break
		}
	}
	if len(roles) == 0 {
		return nil, errors.New("Expected Role 'https://aws.amazon.com/SAML/Attributes/Role'")
	}

	return roles, nil
}

func (a AttributeValue) isRole() bool {
	return a.Name == "https://aws.amazon.com/SAML/Attributes/Role"
}

func (a AttributeValue) arns() []Arn {
	const providerKey = "saml-provider"
	result := []Arn{}
	for _, value := range a.Value {
		parts := strings.Split(value, ",")
		if len(parts) == 2 {
			if strings.Index(parts[0], providerKey) >= 0 {
				result = append(result, Arn{principal: parts[0], role: parts[1]})
			}
			if strings.Index(parts[1], providerKey) >= 0 {
				result = append(result, Arn{principal: parts[1], role: parts[0]})
			}
		}
	}
	return result
}

func lookupAccountAliases(arns []Arn, assertion string, cache *AccountAliasCache, durationSeconds int64) {
	// fix: arn changes do no leave this func
	for i, arn := range arns {
		alias, found := cache.findAlias(arn.role)
		if found {
			arns[i].alias = alias.String()
			continue
		}
		clearCredentials()

		sess := session.New()
		err := assumeRole(sess, &arn, assertion, durationSeconds)

		setCredentials(arn.accessKeyID, arn.secretAccessKey, arn.sessionToken)

		if err != nil {
			journal("assume role for account alias, %s - %v", arn.role, err)
		}

		input := &iam.ListAccountAliasesInput{}
		//todo: now this session has the creds for the newly assumed role
		svc := iam.New(sess)
		result, err := svc.ListAccountAliases(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				journal("Unable to resolve account aliases\n%v - %v - %v\n", aerr.Code(), aerr.Message(), aerr.Error())
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				journal("Unable to resolve account alias")
			}
			// return
			continue
		}

		if len(result.AccountAliases) == 0 {
			journal("No account aliases")
			continue
		}
		cache.add(arns[i].role, result.AccountAliases)
		arns[i].alias = *result.AccountAliases[0]
	}
}
func clearCredentials() {
	os.Unsetenv("AWS_SDK_LOAD_CONFIG")
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_SESSION_TOKEN")
}

func setCredentials(id, secret, token string) {
	os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	os.Setenv("AWS_ACCESS_KEY_ID", id)
	os.Setenv("AWS_SECRET_ACCESS_KEY", secret)
	os.Setenv("AWS_SESSION_TOKEN", token)
}

func (a Arn) String() string {
	return fmt.Sprintf("%s / %s (%s)", a.principal, a.role, a.alias)
}

func (a *Arn) setCredentials(id, secret, token string) {
	a.accessKeyID = id
	a.secretAccessKey = secret
	a.sessionToken = token
}
func (a *Arn) hasCredentials() bool {
	return len(a.accessKeyID) > 0 && len(a.secretAccessKey) > 0 && len(a.sessionToken) > 0
}
func (a *Arn) roleMenu() string {
	return fmt.Sprintf("%s - %s", a.alias, a.role)
}
