# Creating temporary login credentials for AWS CLI with STS

This tool was ported from a python script in the AWS samples s3 bucket [1].

However, it didn't work for me out of the box and used python 3 (by the time it got to me, I got it from a 3rd party) and I only had python 2.7 and I wanted to learn more go.

_For when I can't remember, STS stands for Security Token Service_

## Common errors

### Linux/WSL: `x509: certificate signed by unknown authority`

Make sure the certificates are installed, try `add --no-cache ca-certificates` or `sudo apt-get install ca-certificates` 

## Install (with go tool chain)

```
go get -u github.com/NearlyUnique/awsSts
```

Download the [Current release](/NearlyUnique/awsSts/releases/) for your platform.

[Change log](changelog.md)

## Target Platforms

It works on Windows, I've tested linux (bash on windows) there is no reason it won't work on OSX.

## Usage

Common scenario;
- My STS web login web page is here: `https://sts.domain.company.org/adfs/ls/IdpInitiatedSignOn.aspx?loginToRp=urn:amazon:webservices`
- My other scripts are going to use the `default` AWS profile
- Automatically select the role `arn:aws:iam::123456789:role/my-role`
- Leave running in a state where I can `auto`matically refresh my token with one key press when it expires in an hour

```
awsSts logon --url https://sts.domain.company.org/adfs/ls/IdpInitiatedSignOn.aspx?loginToRp=urn:amazon:webservices --profile default --role arn:aws:iam::123456789:role/my-role
```

`--help` for full details, including details of all parameters that can be read from environment.

## Roadmap
1. Override credential file location via flag
1. Keep running and auto refresh before expiry (optional)
1. Deal with naming of INPUT tags in the login form, the Python sample did some work in this area, I want to improve the guessing ability and allow the user to define it if we can't guess.
1. add command to ease iam user creation
1. add command to rotate iam user secrets
1. Auto upgrade
  - use `runtime.GOOS` and `_VERISON`
  - call `GET https://api.github.com/repos/NearlyUnique/awsSts/releases/latest`
```json
  {
    "tag_name": "0.7",
      "assets": [
        {
          "name": "awsSts-0.7-linux",
          "browser_download_url": "https://some-url"
        }
      ]
  }`
```
  - the `browser_download_url` may give a redirect

## How it works

1. Download the login form, we need the cookies
1. Fill in the user name and password
1. Post form back
1. Parse the response HTML form
1. Find the `SAMLResponse` INPUT element
1. base64 decode it (it's now XML)
1. Extract the Roles, select one
1. Call AWS `AssumeRoleWithSAML`
1. update the credentials ini file with the result

## References

[1] https://s3.amazonaws.com/awsiammedia/public/sample/SAMLAPICLIADFS/samlapi_formauth_adfs3.py
