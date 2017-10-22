# Clisso: CLI Single Sign-On

**WIP Warning! This project is still under development and isn't expected
to work yet.**

*Clisso* allows you to retrieve temporary credentials for cloud
providers and APIs by authenticating with an identity provider (IdP).

The following cloud providers are currently supported:

- [AWS](https://aws.amazon.com/)

The following identity providers are currently supported:

- [OneLogin](https://www.onelogin.com/)

## Installation

1. Inside `$GOPATH/src/github.com/johananl/clisso` run `dep ensure` to install dependencies.
1. Run `go install`. This will put the `clisso` binary in your `$GOPATH/bin` directory.

## Configuration

Create a file called `.clisso.yaml` in your home directory. Following is a
sample configuration:

    clisso:
      defaultApp: dev
    providers:
      onelogin:
        clientSecret: xxxxxxxx
        clientId: xxxxxxxx
        subdomain: example.com
        # Uncomment the following line to specify a default username.
        #username: xxxxxxxx
    apps:
      dev:
        provider: onelogin
        appId: 1234
        principalArn: arn:aws:iam::123456789:saml-provider/another-provider-name
        roleArn: arn:aws:iam::123456789:role/another-role
      prod:
        provider: onelogin
        appId: 5678
        principalArn: arn:aws:iam::123456789:saml-provider/provider-name
        roleArn: arn:aws:iam::123456789:role/a-role


## Usage

Run `clisso get <app-name>` and enter your username, password and OTP
to get temporary credentials.
