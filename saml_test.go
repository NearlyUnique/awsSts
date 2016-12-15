package main

import "testing"

var xxx = `<samlp:Response>
	<Assertion>
		<AttributeStatement>
			<Attribute Name="https://aws.amazon.com/SAML/Attributes/RoleSessionName">
				<AttributeValue>some-user@any-domain</AttributeValue>
			</Attribute>
			<Attribute Name="https://aws.amazon.com/SAML/Attributes/Role">
				<AttributeValue>arn:aws:iam::12345:saml-provider/company,arn:aws:iam::12345:role/some-role-name</AttributeValue>
			</Attribute>
		</AttributeStatement>
	</Assertion>
</samlp:Response>`

func Test_parse_xml(t *testing.T) {
	const expectedRole = "arn:aws:iam::12345678:saml-provider/MyCompanyADFS,arn:aws:iam::12345678:role/mycompany-app-admin"
	a, _ := extractArns([]byte(xxx))
	if a != expectedRole {
		t.Errorf("Missing Role")
	}
}
