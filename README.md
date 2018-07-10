# Clisso: CLI Single Sign-On

**WIP Warning! This project is still under development and isn't expected
to be stable yet.**

Clisso (pronounced `/ˈklIsoʊ/`) allows you to retrieve temporary credentials for cloud providers
by authenticating with an identity provider (IdP).

The following cloud providers are currently supported:

- [AWS][1]

The following identity providers are currently supported:

- [OneLogin][2]
- [Okta][3]

## Installation

### Using a Pre-Compiled Binary

The easiest way to use Clisso is to download a pre-compiled binary for your platform. To do so,
perform the following:

1. Go to the [latest release][4] on the releases page.
1. Download the ZIP file corresponding to your platform and architecture.
1. Unzip the binary.
1. Rename the binary using `mv clisso-<platform>-<arch> clisso`.
1. Move the binary to a place under your path.

### Building from Source

Clisso uses [dep][5] for dependency management. You will need it to install dependencies. You will
also need Go to compile the code, and Git which is used by `go get`.

- Git
- Go
- dep

To build Clisso from source, do the following:

1. Download the source code using `go get -d github.com/allcloud-io/clisso`.
1. `cd` to `$GOPATH/src/github.com/allcloud-io/clisso`.
1. Run `dep ensure` to install the dependencies.
1. Run `make` to build the binary.
1. Run `make install` to put the binary in your `$PATH`.
1. Run `make clean` to clean up after the build.

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

## Caveats and Limitations

- No Windows support.
- No support for Okta applications with MFA enabled **at the application level**.
- No support for IAM role selection.

[1]: https://aws.amazon.com/
[2]: https://www.onelogin.com/
[3]: https://www.okta.com/
[4]: https://github.com/allcloud-io/clisso/releases/latest
[5]: https://github.com/golang/dep
