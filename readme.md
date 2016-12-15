# Creating temporary login credentials for AWS CLI

This tool was ported from a python script in the AWS samples s3 bucket [1].

However, it didn't work for me out of the box and used python 3 (by the time it got to me, I got it from a 3rd party) and I only had python 2.7 and I wanted to learn more go.

# Install (with go tool chain)

```
go install -u github.com/NearlyUnique/awsSts
```

# Target Platforms

It works on Windows, there is no reason it won't work on Linux and OSX.

# Usage

```
SET AWSSTS_URL=https://sts.domain.company.org/adfs/ls/IdpInitiatedSignOn.aspx?loginToRp=urn:amazon:webservices
SET AWSSTS_USER=my-username@domain.company.org
awsSts --profile default
```

`--profile` is optional, default is `saml` but it is useful to seitch the default

# Road Map
1. Code review and tidy up
1. Override credential file location via flag
1. Keep running and auto refresh before expiry (optional)
1. Deal with naming of INPUT tags in the login form, the Python sample did some work in this area, I want to improve the guessing ability and allow the user to define it if we can't guess.
1. If there are no features in teh aws cli, add features here to switch profiles round, eg. named to default and vice versa, delete, add default region

# How it works

1. Download the login form, we need the cookies
1. Fill in the user name and password
1. Post form back
1. Parse the HTML form
1. Find the `SAMLResponse` element
1. base64 decode it (it's XML)
1. Extract the Roles, select one
1. Call AWS `AssumeRoleWithSAML`
1. update the credentials ini file


[1] https://s3.amazonaws.com/awsiammedia/public/sample/SAMLAPICLIADFS/samlapi_formauth_adfs3.py