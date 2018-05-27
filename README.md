# Clisso: CLI Single Sign-On

**WIP Warning! This project is still under development and isn't expected
to be stable yet.**

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

    apps:
      sandbox:
        appid: "123456"
        principalarn: arn:aws:iam::123456789:saml-provider/OneLoginDev
        provider: onelogin-dev
        rolearn: arn:aws:iam::123456789:role/OneLoginDev-SSO
    global:
      credentialsfilepath: ~/.aws/credentials
    providers:
      onelogin-dev:
        clientid: xxxxxxxx
        clientsecret: xxxxxxxx
        subdomain: example
        type: onelogin
        # Uncomment the following line to specify a default username.
        # username: xxxxxxxx

## Usage

Run `clisso get <app-name>` and enter your username, password and OTP
to get temporary credentials.
