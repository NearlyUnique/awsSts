package main

import "testing"

var xxx = `<samlp:Response>
	<Assertion>
		<AttributeStatement>
			<Attribute Name="https://aws.amazon.com/SAML/Attributes/RoleSessionName">
				<AttributeValue>some-user@any-domain</AttributeValue>
			</Attribute>
			<Attribute Name="https://aws.amazon.com/SAML/Attributes/Role">
				<AttributeValue>arn:aws:iam::12345678:saml-provider/MyCompanyADFS,arn:aws:iam::12345678:role/mycompany-app-admin</AttributeValue>
			</Attribute>
		</AttributeStatement>
	</Assertion>
</samlp:Response>`

func Test_parse_xml(t *testing.T) {
	const expectedRole = "arn:aws:iam::12345678:role/mycompany-app-admin"
	const expectedPrincipal = "arn:aws:iam::12345678:saml-provider/MyCompanyADFS"
	p, r := extractArns([]byte(xxx))
	if p != expectedPrincipal {
		t.Errorf("Incorrect principal\ngot: %s\nexp :%s", p, expectedPrincipal)
	}
	if r != expectedRole {
		t.Errorf("Incorrect principal\ngot: %s\nexp :%s", r, expectedRole)
	}
}
