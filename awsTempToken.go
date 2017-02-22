package main

import (
	"encoding/base64"
	"encoding/xml"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/vito/go-interact/interact"
)

func getAwsTempToken(role string, resp *http.Response) (*Arn, string) {
	xml, assertion, err := getSaml(resp)
	exitErr(err, "Unable to read SAML")

	arns, err := extractArns(xml)
	exitErr(err, "Failed to parse arns")

	arn, err := selectRole(role, arns)
	exitErr(err, "Failed to select role")
	return arn, assertion
}
func getSaml(resp *http.Response) (xml []byte, assertion string, err error) {
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, "", errors.Wrap(err, "Prepare login form for query INPUT elements")
	}
	rawHTML, _ := doc.Html()
	dumpFile("login-response-", []byte(rawHTML))

	inputs := []string{}
	doc.Find("input").Each(func(i int, s *goquery.Selection) {
		if err != nil {
			return
		}
		name := attrOrEmpty(s, "name")
		if name == "SAMLResponse" {
			assertion = attrOrEmpty(s, "value")
			xml, err = base64.StdEncoding.DecodeString(assertion)
		}
		inputs = append(inputs, name)
	})
	if len(xml) == 0 {
		msg := loginErrorText(doc)
		if len(msg) == 0 {
			msg = "Failed to find SAML xml"
		}
		return nil, "", errors.New(msg)
	}
	return xml, assertion, errors.Wrap(err, "Base64 decoding 'assertion' XML")
}
func extractArns(raw []byte) ([]Arn, error) {
	response := Response{}
	err := xml.Unmarshal(raw, &response)
	dumpFile("saml-xml-", raw)
	if err != nil {
		return nil, errors.Wrap(err, "Reading SAML XML")
	}
	roles := []Arn{}
	for _, a := range response.Assertion {
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
func selectRole(defaultRole string, roles []Arn) (*Arn, error) {
	if len(roles) == 1 {
		log("Using Role %q\n", roles[0].role)
		return &roles[0], nil
	}
	var list []interact.Choice
	list = append(list, interact.Choice{Display: "Quit", Value: 0})

	for i, a := range roles {
		if a.role == defaultRole {
			log("Default role found %s\n", a.role)
			return &a, nil
		}
		list = append(list, interact.Choice{Display: a.role, Value: i + 1})
	}
	if len(defaultRole) > 0 {
		log("Default role not found %s", defaultRole)
	}

	choice := 1
	err := interact.NewInteraction(
		"Choose a role",
		list...,
	).Resolve(&choice)
	if choice == 0 {
		log("Canceled by user")
		os.Exit(0)
	}
	if err != nil {
		return nil, errors.Wrap(err, "No choice")
	}
	return &roles[choice-1], nil
}
func loginErrorText(doc *goquery.Document) string {
	for _, id := range []string{"#errorText", "#expiredNotification"} {
		sel := doc.Find(id)
		if sel != nil {
			msg := strings.Trim(sel.Text(), " \t\r\n")
			if len(msg) > 0 {
				return msg
			}
		}
	}
	return ""
}
func loginDetails(user, pass string) (form map[string]string, err error) {
	if len(user) == 0 {
		user, err = getUsername()
	} else {
		log("Username (from %s)='%s'\n", userEnv, user)
	}
	if err == nil && len(pass) == 0 {
		pass, err = getPassword()
	} else {
		log("Password (from %s), length='%d'\n", passEnv, len(pass))
	}
	return map[string]string{
		keyUsername: user,
		keyPassword: pass,
	}, err //error already wrapped
}
func getUsername() (string, error) {
	var username = ""
	err := interact.NewInteraction("Username").Resolve(interact.Required(&username))
	if err != nil {
		return "", errors.Wrap(err, "Reading username")
	}
	return string(username), nil
}
func getPassword() (string, error) {
	var password interact.Password
	err := interact.NewInteraction("Password").Resolve(interact.Required(&password))
	if err != nil {
		return "", errors.Wrap(err, "Reading password")
	}
	return string(password), nil
}
func attrOrEmpty(s *goquery.Selection, name string) string {
	if r, ok := s.Attr(name); ok {
		return r
	}
	return ""
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
				result = append(result, Arn{parts[0], parts[1]})
			}
			if strings.Index(parts[1], providerKey) >= 0 {
				result = append(result, Arn{parts[1], parts[0]})
			}
		}
	}
	return result
}
