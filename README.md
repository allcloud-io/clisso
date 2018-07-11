# Clisso: CLI Single Sign-On

**WIP Warning! This project is still under development and isn't expected
to be stable yet.**

Clisso (pronounced `/ˈklIsoʊ/`) allows you to retrieve temporary credentials for cloud platforms
by authenticating with an identity provider (IdP).

The following identity providers are currently supported:

- [OneLogin][2]
- [Okta][3]

The following cloud platforms are currently supported:

- [AWS][1]

Clisso uses the [SAML][7] standard to authenticate users.

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
also need **Go** to compile the code, and **Git** which is used by `go get`.

To build Clisso from source, do the following:

1. Download the source code using `go get -d github.com/allcloud-io/clisso`.
1. `cd` to `$GOPATH/src/github.com/allcloud-io/clisso`.
1. Run `dep ensure` to install the dependencies.
1. Run `make` to build the binary.
1. Run `make install` to put the binary in your `$PATH`.
1. Run `make clean` to clean up after the build.

## Configuration

Clisso stores configuration in a file called `.clisso.yaml` under the user's home directory. You
may specify a different config file using the `-c` flag.

>NOTE: It is recommended to use the `clisso` command to manage the config file, however you may
>also edit the file manually. The file is in YAML format.

## Usage

Clisso has the following commands:

    $ ./clisso
    Usage:
    clisso [command]

    Available Commands:
    apps        Manage apps
    get         Get temporary credentials for an app
    help        Help about any command
    providers   Manage providers
    version     Show version info

    Flags:
    -c, --config string   config file (default is $HOME/.clisso.yaml)
    -h, --help            help for clisso

    Use "clisso [command] --help" for more information about a command.

In order to use Clisso you will have to configure at least one *provider* and one *app*. A provider
represents an identity provider against which Clisso authenticates. An app represents an account
on a cloud platform such as AWS, for which Clisso retrieves credentials.

### Configuring Providers

#### OneLogin

To create a OneLogin identity provider, use the following command:

    clisso providers create onelogin my-provider \
        --client-id myid \
        --client-secret mysecret \
        --subdomain mycompany \
        --username user@mycompany.com

The example above creates a OneLogin identity provider configuration for Clisso, with the name
`my-provider`.

The `--client-id` and `--client-secret` flags are OneLogin API credentials. You may follow the
instructions [here][8] to obtain them. OneLogin requires using static credentials even for
**attempting authentication**, and for that reason Clisso needs them. Please be sure to select
**Authentication Only** when generating the credentials. Higher-level permissions aren't used by
Clisso and will only pose a security risk when stored at a client machine.

The `--subdomain` flag is the subdomain of your OneLogin account. You can see it in the URL when
logging in to OneLogin. For example, if you log in to OneLogin using `mycompany.onelogin.com`, use
`--subdomain mycompany`.

The `--username` flag is optional, and allows Clisso to always use the given value as the OneLogin
username when retrieving credentials for apps which use this provider. Omitting this flag will make
Clisso prompt for a username every time.

#### Okta

To create an Okta identity provider, use the following command:

    clisso providers create okta my-provider \
        --base-url https://mycompany.okta.com \
        --username user@mycompany.com

The example above creates an Okta identity provider configuration for Clisso, with the name
`my-provider`.

The `--base-url` flag is your Okta base URL. You can see it in the URL when logging in to Okta.
Please specify a full URL in one of the following formats:

- `https://your-subdomain.okta.com` if you have an enterprise Okta account.
- `https://your-subdomain.oktapreview.com` if you have a developer Okta account.

The `--username` flag is optional, and allows Clisso to always use the given value as the Okta
username when retrieving credentials for apps which use this provider. Omitting this flag will make
Clisso prompt for a username every time.

### Configuring Apps

#### OneLogin

To create a OneLogin app, use the following command:

    clisso apps create onelogin my-app \
        --provider my-provider \
        --app-id 12345 \
        --principal-arn arn:aws:iam::123456789012:saml-provider/OneLogin \
        --role-arn arn:aws:iam::123456789012:role/OneLoginSSO

The example above creates a OneLogin app configuration for Clisso, with the name `my-app`.

The `--provider` flag is the name of a provider which already exists in the config file.

The `--app-id` flag is the OneLogin app ID. This ID can be retrieved using the OneLogin admin
interface or the OneLogin API. Unfortunately, the OneLogin API doesn't allow obtaining app IDs
without storing sensitive, high-level permissions on the client machine. For that reason we have to
manually configure the app ID for every app.

>NOTE: The ID seen in the browser URL when visiting a OneLogin app as a user is **NOT** the app ID.
>Only a OneLogin administrator can obtain an app ID.

The `--principal-arn` is the ARN of the [identity provider][9] that was created on AWS IAM for the
OneLogin integration.

The `--role-arn` is the ARN of the IAM role that should be assumed following a successful
authentication against OneLogin.

#### Okta

TODO

### Obtaining Credentials

To obtain temporary credentials for an app, use the following command:

    $ clisso get <app-name>

By default, Clisso will try to store the credentials in the [shared credentials file][6] of the AWS
CLI. To save the credentials to a different file, use the `-w` flag.

To print the credentials to the shell instead of storing them in a file, use the `-s` flag.

## Caveats and Limitations

- No support for Okta applications with MFA enabled **at the application level**.
- No support for IAM role selection.

[1]: https://aws.amazon.com/
[2]: https://www.onelogin.com/
[3]: https://www.okta.com/
[4]: https://github.com/allcloud-io/clisso/releases/latest
[5]: https://github.com/golang/dep
[6]: https://docs.aws.amazon.com/cli/latest/userguide/cli-config-files.html
[7]: https://en.wikipedia.org/wiki/Security_Assertion_Markup_Language
[8]: https://developers.onelogin.com/api-docs/1/getting-started/working-with-api-credentials
[9]: https://onelogin.service-now.com/support?id=kb_article&sys_id=de999903db109700d5505eea4b961966
