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

### Obtaining Credentials

To obtain temporary credentials for an app, use the following command:

    $ clisso get <app-name>

By default, Clisso will try to store the credentials in the [shared credentials file][6] of the AWS
CLI. To save the credentials to a different file, use the `-w` flag.

To print the credentials to the shell instead of storing them in a file, use the `-s` flag.

### Configuring Providers

TODO

### Configuring Apps

TODO

## Caveats and Limitations

- No support for Okta applications with MFA enabled **at the application level**.
- No support for IAM role selection.

[1]: https://aws.amazon.com/
[2]: https://www.onelogin.com/
[3]: https://www.okta.com/
[4]: https://github.com/allcloud-io/clisso/releases/latest
[5]: https://github.com/golang/dep
[6]: https://docs.aws.amazon.com/cli/latest/userguide/cli-config-files.html
