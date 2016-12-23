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
	roles, err := extractArns([]byte(saml))
	ok(t, err)
	equals(t, len(roles), 1)
	equals(t, expectedPrincipal, roles[0].principal)
	equals(t, expectedRole, roles[0].role)
}

func Test_parse_xml_multiple_roles(t *testing.T) {
	const expectedRole1 = "arn:aws:iam::12345678:role/mycompany-app-admin1"
	const expectedPrincipal1 = "arn:aws:iam::12345678:saml-provider/MyCompanyADFS2"
	const expectedRole2 = "arn:aws:iam::12345678:role/mycompany-app-admin2"
	const expectedPrincipal2 = "arn:aws:iam::12345678:saml-provider/MyCompanyADFS2"
	saml := makeSamlXML(
		combineRoleAttrs(expectedRole1, expectedPrincipal1),
		combineRoleAttrs(expectedRole2, expectedPrincipal2))
	roles, err := extractArns([]byte(saml))
	ok(t, err)
	equals(t, len(roles), 2)
	equals(t, expectedPrincipal1, roles[0].principal)
	equals(t, expectedRole1, roles[0].role)
	equals(t, expectedPrincipal2, roles[1].principal)
	equals(t, expectedRole2, roles[1].role)
}
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
