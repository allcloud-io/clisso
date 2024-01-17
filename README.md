# Clisso: CLI Single Sign-On

[![Coverage Status](https://coveralls.io/repos/github/allcloud-io/clisso/badge.svg)](https://coveralls.io/github/allcloud-io/clisso)

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

Clisso supports **macOS**, **Linux** and **Windows**.

### Installing Using Homebrew

To install Clisso using Homebrew, run the following commands:

    brew tap allcloud-io/tools
    brew install clisso

To update Clisso to the latest release, run the following command:

    brew upgrade clisso

### Building from Source

#### Requirements

- Go `1.12` or above
- Git
- Make

#### Building

To build Clisso from source, do the following:

```
# Get the source
git clone github.com/allcloud-io/clisso

# Build the binary
cd clisso
go build

# Install the binary in $GOPATH/bin
go install

# Clean up
go clean
```

## Configuration

Clisso stores configuration in a file called `.clisso.yaml` under the user's home directory. You
may specify a different config file using the `-c` flag.

>NOTE: It is recommended to use the `clisso` command to manage the config file, however you may
>also edit the file manually. The file is in YAML format. You may find a sample config file
>[here][11].

## Usage

Clisso has the following commands:

    $ ./clisso
    Usage:
        clisso [command]

    Available Commands:
        apps        Manage apps
        completion  Generate the autocompletion script for the specified shell
        get         Get temporary credentials for an app
        help        Help about any command
        providers   Manage providers
        status      Show active (non-expired) credentials

    Flags:
        -c, --config string      config file (default is $HOME/.clisso.yaml)
        -h, --help               help for clisso
            --log-level string   set log level to trace, debug, info, warn, error, fatal or panic (default "info")
        -v, --version            version for clisso

    Use "clisso [command] --help" for more information about a command.

In order to use Clisso you will have to configure at least one *provider* and one *app*. A provider
represents an identity provider against which Clisso authenticates. An app represents an account
on a cloud platform such as AWS, for which Clisso retrieves credentials.

### Listing Providers

To list the existing providers on Clisso, use the following command:

    clisso providers ls

Following is a sample output:

    okta-prod
    onelogin-dev
    onelogin-prod

### Listing Apps

To list the existing apps on Clisso, use the following command:

    clisso apps ls

Following is a sample output:

      dev-account
    * prod-account

The app marked with an asterisk is [selected](#selecting-an-app).

### Creating Providers

#### OneLogin

To create a OneLogin identity provider, use the following command:

    clisso providers create onelogin my-provider \
        --client-id myid \
        --client-secret mysecret \
        --subdomain mycompany \
        --username user@mycompany.com \
        --region US \
        --duration 14400 \
        --arn arn:aws:iam::123456789012:role/Worker

The example above creates a OneLogin identity provider configuration for Clisso, with the name
`my-provider`.

The `--client-id` and `--client-secret` flags are OneLogin API credentials. You may follow the
instructions [here][8] to obtain them. OneLogin requires using static credentials even for
**attempting authentication**, and for that reason Clisso needs them. Please be sure to select
**Authentication Only** when generating the credentials. Higher-level permissions aren't used by
Clisso and will only pose a security risk when stored at a client machine. You might have to open
a ticket with your OneLogin administrator to obtain these credentials as administrator privileges
are required.

The `--subdomain` flag is the subdomain of your OneLogin account. You can see it in the URL when
logging in to OneLogin. For example, if you log in to OneLogin using `mycompany.onelogin.com`, use
`--subdomain mycompany`.

The `--username` flag is optional, and allows Clisso to always use the given value as the OneLogin
username when retrieving credentials for apps which use this provider. Omitting this flag will make
Clisso prompt for a username every time.

The `--duration` flag is optional. If specified, sessions will be assumed with the provided
duration, in seconds, instead of the default of 3600 (1 hour). Valid values are between 3600 and
43200 seconds. The [max session duration][12] has be equal to or lower than what is configured on
the role in AWS. If a longer session time is requested than what is configured on the AWS role,
Clisso will fallback to a duration of 3600. The default duration specified for the provider can be
overridden on a per-app basis (see below).

The `--arn` flag is optional. If specified, it will not prompt for a choice of roles presented
from the list of available AWS accounts/roles. This makes it easy to run `clisso get my-app`
and get the correct account/role.

#### Okta

To create an Okta identity provider, use the following command:

    clisso providers create okta my-provider \
        --base-url https://mycompany.okta.com \
        --username user@mycompany.com \
        --duration 14400

The example above creates an Okta identity provider configuration for Clisso, with the name
`my-provider`.

The `--base-url` flag is your Okta base URL. You can see it in the URL when logging in to Okta.
Please specify a full URL in one of the following formats:

- `https://your-subdomain.okta.com` if you have an enterprise Okta account.
- `https://your-subdomain.oktapreview.com` if you have a developer Okta account.

The `--username` flag is optional, and allows Clisso to always use the given value as the Okta
username when retrieving credentials for apps which use this provider. Omitting this flag will make
Clisso prompt for a username every time.

The `--duration` flag is optional. If specified, sessions will be assumed with the provided
duration, in seconds, instead of the default of 3600 (1 hour). Valid values are between 3600 and
43200 seconds. The [max session duration][12] has be equal to or lower than what is configured on
the role in AWS. If a longer session time is requested than what is configured on the AWS role,
Clisso will fallback to a duration of 3600. The default duration specified for the provider can be
overridden on a per-app basis (see below).

### Deleting Providers

Deleting providers using the `clisso` command isn't currently supported. To delete a provider,
remove its configuration from the config file.

### Creating Apps

#### OneLogin

To create a OneLogin app, use the following command:

    clisso apps create onelogin my-app \
        --provider my-provider \
        --app-id 12345 \
        --duration 3600

The example above creates a OneLogin app configuration for Clisso, with the name `my-app`.

The `--provider` flag is the name of a provider which already exists in the config file.

The `--app-id` flag is the OneLogin app ID. This ID can be retrieved using the OneLogin admin
interface or the OneLogin API. Unfortunately, the OneLogin API doesn't allow obtaining app IDs
without storing sensitive, high-level permissions on the client machine. For that reason we have to
manually configure the app ID for every app.

>NOTE: The ID seen in the browser URL when visiting a OneLogin app as a user is **NOT** the app ID.
>Only a OneLogin administrator can obtain an app ID.

The `--duration` flag is optional and defaults to the value set at the provider level. Valid values
are between 3600 and 43200 seconds. Can be used to raise or lower the session duration for an
individual app. The [max session duration][12] has be equal to or lower than what is configured on
the role in AWS. The default maximum is 3600 seconds. If the requested duration exceeds the
configured maximum Clisso will fallback to 3600 seconds.

#### Okta

To create an Okta app, use the following command:

    clisso apps create okta my-app \
        --provider my-provider \
        --url https://mycompany.okta.com/home/amazon_aws/xxxxxxxxxxxxxxxxxxxx/137 \
        --duration 3600

The example above creates an Okta app configuration for Clisso, with the name `my-app`.

The `--provider` flag is the name of a provider which already exists in the config file.

The `--url` flag is the app's **embed link**. This can be retrieved as an Okta user by examining
the URL of an app on the Okta web UI. The same can also be retrieved as an administrator by
clicking an app in the **Applications** view. The embed link is on the **General** tab.

>NOTE: An Okta embed link must not contain an HTTP query, only the base URL. For AWS apps, the link
should end with `/137`.

The `--duration` flag is optional and defaults to the value set at the provider level. Valid values
are between 3600 and 43200 seconds. Can be used to raise or lower the session duration for an
individual app. The [max session duration][12] has be equal to or lower than what is configured on
the role in AWS. The default maximum is 3600 seconds. If the requested duration exceeds the
configured maximum Clisso will fallback to 3600 seconds.

### Deleting Apps

For deleting apps, use the following command:

    clisso apps delete my-app

Deletion of an app will remove its configuration from the config file. You can also do it manually
by editing the config file.

### Obtaining Credentials

To obtain temporary credentials for an app, use the following command:

    clisso get my-app

The example above will obtain credentials for an app named `my-app`. Type your credentials for the
relevant identity provider. If multi-factor authentication is enabled on your account, you will be
asked in addition for a one-time password.

By default, Clisso will store the credentials in the [shared credentials file][6] of the AWS CLI
with the app's name as the [profile name][10]. You can use the temporary credentials by specifying
the profile name as an argument to the AWS CLI (`--profile my-profile`), by setting the
`AWS_PROFILE` environment variable or by configuring any AWS SDK to use the profile.

To save the credentials to a custom file, use the `--output` flag with a custom path. For example:

    clisso get my-app --output /path/to/credentials

To print the credentials to the shell instead of storing them in a file, use the `--output environment` flag. This
will output shell commands which can be pasted in any shell to use the credentials.

To select a specific MFA device by name instead of choosing from a list, use the `-m` flag. The 
configuration field `global.mfa-device` may also be set.

### Running as `credential_process`

AWS CLI v2 introduced the `credential_process` feature which allows you to use an external command to obtain temporal credentials.
Clisso can be used as a `credential_process` command by setting the `--output credential_process` flag. For example:

    clisso get my-app --output credential_process

You can use this by adding the following to your `~/.aws/credentials` file:

```ini
[my-app]
credential_process = clisso get my-app --output credential_process
```

> **IMPORTANT**: If `clisso get my-app --output credential_process` prompts for any input, the `credential_process` will not work as expected. Make sure to configure Clisso to not prompt for any input (Store the password in the key chain, use push MFA).

Alternatively you can run the following command to configure all Apps for use with `credential_process`:

```bash
clisso cp configure
```

The AWS SDK does not cache any credentials obtained using `credential_process`. This means that every time you use the profile, Clisso will be called to obtain new credentials. If you want to cache the credentials, you can use the `--cache` flag. For example:

```ini
[my-app]
credential_process = clisso get my-app --output credential_process --cache
```

Alternatively you can set it in the `~/.clisso.yaml` file:

```yaml
global:
  cache:
    enable: true
```

#### Temporarily Disabling Credential Process Functionality

Different processes on your system might continue using AWS Profiles configured for use with Clisso. To temporarily disable the `credential_process` functionality, you can use the `clisso cp` submenu. For example:

```bash
clisso cp disable # to disable
clisso cp enable # to enable
clisso cp status # to check the status
```

If you disable the `credential_process` functionality, all refreshes will be disabled. While cached credentials will still be used, new credentials will not be fetched. This can be useful if you lock your computer with an active, e.g., VSCode session with CodeCommit. If you wouldn't disable the `credential_process` functionality, the VSCode would constantly trigger new credential requests to refresh the remote CodeCommit repository.

If you want to check the status programmatically, you can use the exit code of the `clisso cp status` command. If the exit code is `0`, the `credential_process` functionality is enabled. If the exit code is `1`, the `credential_process` functionality is disabled.

### Storing the password in the key chain

> WARNING: Storing the password without having MFA enabled is a security risk. It allows anyone
> to assume your roles who has access to your computer.

Storing a password for a provider is as simple as running:

    clisso providers passwd my-provider

### Selecting an App

You can **select** an app by using the following command:

    clisso apps select my-app

You can get credentials for the currently-selected app by simply running `clisso get`, without
specifying an app name. The currently-selected app will have an asterisk near its name when listing
apps using `clisso apps ls`.

## AWS STS Regional Endpoint

AWS recommends using [regional STS endpoints](https://docs.aws.amazon.com/sdkref/latest/guide/feature-sts-regionalized-endpoints.html) instead of the default Global endpoint when requesting a token.

To use a regional endpoint, specify the region via the `global.aws-region` field in the config file. A per app configuration using `apps.<app>.aws-region` is also possible.

## YubiKey Autodetection

YubiKey Autodetection is available for the OneLogin provider. To enable this feature set the `global.autodetect-yubikey` field to `true`. A per app configuration using `apps.<app>.autodetect-yubikey` is also available. Clisso will look at attached USB devices and automatically select the YubiKey as an MFA device if it is available. 

## Caveats and Limitations

- No support for Okta applications with MFA enabled **at the application level**.

## Troubleshooting

### Clisso is not working

Clisso logs to `stderr` by default. To enable more detailed logging, set the `--log-level` flag to `debug` or `trace`. With `trace` log level, sensitive information will be logged.

### Creating a trace log

If you run into issues, you can create a trace log by setting the `--log-level` flag to `trace`. This will create a file called `.clisso.log` your home directory. You can alter the location of the log file by setting the `--log-file` flag. The below example will create a trace log in the current directory in
a file called `trace.log`.

    clisso --log-level trace --log-file trace.log get my-app

Alternatively, you can configure logging via the config file. The below example will create a trace log in your home directory in a file called `clisso.log`.

```yaml
global:
  log:
    level: trace
    file: ~/clisso.log
```

When attaching the log file to an issue, please make sure to remove any sensitive information.

### Storing passwords is not working

`dbus: couldn't determine address of session bus` This behavior has been [observed][13] on Ubuntu 20.04 WSL.
Simply running `sudo systemd-machine-id-setup` gets you past the initial missing machine id setup.

`failed to unlock correct collection '/org/freedesktop/secrets/collection/login'`,
`The name org.freedesktop.secrets was not provided by any .service files` Check that you have a working
keychain setup. On headless systems like WSL this might not be easy to archive. Installing `gnome-keyring`
along with the proper DBus setup is required. During tests adding the below to the `~/.bashrc` on Ubuntu 20.04 WSL
was enough.

```bash
 if [ "$DBUS_SESSION_BUS_ADDRESS" = "" ]; then
    exec dbus-run-session -- bash;
else
    eval $(echo "$(/lib/cryptsetup/askpass 'Password: ')" | gnome-keyring-daemon --unlock);
fi
```
## Contributing

TODO

## License

This code is released under the Mozilla Public License 2.0 License. Please see [LICENSE](LICENSE) and [NOTICE](NOTICE) for more details.

Copyright &copy; 2017-2023 AllCloud

[1]: https://aws.amazon.com/
[2]: https://www.onelogin.com/
[3]: https://www.okta.com/
[4]: https://github.com/allcloud-io/clisso/releases/latest
[5]: https://github.com/golang/dep
[6]: https://docs.aws.amazon.com/cli/latest/userguide/cli-config-files.html
[7]: https://en.wikipedia.org/wiki/Security_Assertion_Markup_Language
[8]: https://developers.onelogin.com/api-docs/1/getting-started/working-with-api-credentials
[9]: https://onelogin.service-now.com/support?id=kb_article&sys_id=de999903db109700d5505eea4b961966
[10]: https://docs.aws.amazon.com/cli/latest/userguide/cli-multiple-profiles.html
[11]: sample_config.yaml
[12]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use.html#id_roles_use_view-role-max-session
[13]: https://github.com/Versent/saml2aws/issues/436
[14]: https://github.com/zalando/go-keyring/issues/48
