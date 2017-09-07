package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
)

func assumeRole(sess *session.Session, arn *Arn, assertion string) error {
	svc := sts.New(sess)
	params := &sts.AssumeRoleWithSAMLInput{
		PrincipalArn:  aws.String(arn.principal),
		RoleArn:       aws.String(arn.role),
		SAMLAssertion: aws.String(assertion),
	}
	resp, err := svc.AssumeRoleWithSAML(params)
	if err != nil {
		return errors.Wrapf(err, "Unable to assume role:%s", arn.role)
	}
	c := resp.Credentials
	arn.setCredentials(*c.AccessKeyId, *c.SecretAccessKey, *c.SessionToken)
	return nil
}

// type profile struct {
// 	credentials.Value
// }

// func createProfile(key, secret, session string) *profile {
// 	p := profile{}
// 	p.Value = credentials.Value{
// 		AccessKeyID:     key,
// 		ProviderName:    "Sts",
// 		SecretAccessKey: key,
// 		SessionToken:    session,
// 	}
// 	return &p
// }

// func (p *profile) Retrieve() (credentials.Value, error) {
// 	return p.Value, nil
// }
// func (p *profile) IsExpired() bool { return false }
