package main

import (
	"fmt"
	"testing"
)

var XMLTemplate = `<samlp:Response>
	<Assertion>
		<AttributeStatement>
			<Attribute Name="https://aws.amazon.com/SAML/Attributes/RoleSessionName">
				<AttributeValue>some-user@any-domain</AttributeValue>
			</Attribute>
			<Attribute Name="https://aws.amazon.com/SAML/Attributes/Role">
				%s
			</Attribute>
		</AttributeStatement>
	</Assertion>
</samlp:Response>`

func Test_parse_xml_single_role(t *testing.T) {
	const expectedRole = "arn:aws:iam::12345678:role/mycompany-app-admin"
	const expectedPrincipal = "arn:aws:iam::12345678:saml-provider/MyCompanyADFS"
	saml := makeSamlXML(combineRoleAttrs(expectedRole, expectedPrincipal))
	arn, err := extractArns([]byte(saml))
	ok(t, err)
	equals(t, expectedPrincipal, arn.principal)
	equals(t, expectedRole, arn.role)
}

// func Test_parse_xml_multiple_roles(t *testing.T) {
// 	const expectedRole = "arn:aws:iam::12345678:role/mycompany-app-admin"
// 	const expectedPrincipal = "arn:aws:iam::12345678:saml-provider/MyCompanyADFS"
// 	saml := makeSamlXML(combineRoleAttrs(expectedRole, expectedPrincipal))
// 	p, r := extractArns([]byte(saml))
// 	if p != expectedPrincipal {
// 		t.Errorf("Incorrect principal\ngot: %s\nexp :%s", p, expectedPrincipal)
// 	}
// 	if r != expectedRole {
// 		t.Errorf("Incorrect principal\ngot: %s\nexp :%s", r, expectedRole)
// 	}
// }
func combineRoleAttrs(a, b string) string {
	return a + "," + b
}
func makeSamlXML(roleXML ...string) string {
	roles := ""
	for _, r := range roleXML {
		roles += "<AttributeValue>" + r + "</AttributeValue>"
	}
	return fmt.Sprintf(XMLTemplate, roles)
}
