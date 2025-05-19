# Metadata Server

[![Go Report Card](https://goreportcard.com/badge/github.com/minherz/metadataserver)](https://goreportcard.com/report/github.com/minherz/metadataserver)
[![Build](https://github.com/minherz/metadataserver/actions/workflows/go.yaml/badge.svg)](https://github.com/minherz/metadataserver/actions/workflows/go.yaml)
[![codecov](https://codecov.io/gh/minherz/metadataserver/branch/main/graph/badge.svg)](https://codecov.io/gh/minherz/metadataserver)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/minherz/metadataserver)](https://pkg.go.dev/github.com/minherz/metadataserver)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/minherz/metadataserver?label=go-version)
![Repo license](https://badgen.net/badge/license/Apache%202.0/blue)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fminherz%2Fmetadataserver.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fminherz%2Fmetadataserver?ref=badge_shield)

This is a simulation of the metadata server that run on cloud environments of such providers as Amazon, Google or Azure.
This server is intended to assist in local debugging of applications that are intended for run in cloud environments and make use of the environment's metadata server.

The default configuration of the metadataserver package sets up the following endpoint:

* `http://169.254.169.254/computeMetadata/v1`

All other endpoints are served by appending the path of the metadata to the default endpoint path.

> [!NOTE]
> Currently metadataserver does not support HTTPS endpoints

## Use the package

To use the package do the following:

1. Import the package into your code:

   ```go
   import "github.com/minherz/metadataserver"
   ```

2. Instantiate the server:

   ```go
   ms, err := metadataserver.New(metadataserver.WithConfigFile("path/to/config/file"))
   ```

   See other [options] for more configurations.

3. Start the server:

   ```go
   err := ms.Start(context.Background())
   ```

4. Stop the server:

   ```go
   err := ms.Stop(context.Background())
   ```

### Using for unit testing

This package can be used for local unit testing of the code that uses metadata server.

> [!IMPORTANT]
> If your code sends requests to the metadata server using IP address, make sure that this IP address is reachable.
> Reference to [Metadata server IP address](#metadata-server-ip-address) section for more details.
>
> If your code sends requests to the metadata server using hostname, [edit hostname file](https://www.howtogeek.com/27350/beginner-geek-how-to-edit-your-hosts-file/) to link the hostname to localhost.

In your test file start and stop the server as shown above.

#### Test function example

The following example configures the metadata server to run on interloop interface, listening at port 80 and serving requests at the following two endpoints:

* computeMetadata/v1/project-id
* computeMetadata/v1/instance/zone

```go
package service_test

import (
 "context"
 "testing"

 "github.com/org/project/service"
 "github.com/minherz/metadataserver"
)

func TestMyService(t *testing.T) {
 if testing.Short() {
  t.Skip()
 }
 tests := []struct {
  name  string
  // rest of input/want/expected data
 }{
  name: "test1",
 }}

 s, err := metadataserver.New(
    metadataserver.WithAddress("0.0.0.0"),
    metadataserver.WithPort(80),
    metadataserver.WithHandlers(
      map[string]metadataserver.Metadata {
        "project-id": func() string { return "your-test-project" },
        "instance/zone": func() string { return "us-central1" },
      }))
 if err != nil {
  t.Errorf("expected no errors, got: %v", err)
 }
 err = s.Start()
 if err != nil {
  t.Errorf("expected no errors, got: %v", err)
 }
 defer s.Stop()
 for _, test := range tests {
  t.Run(test.name, func(t *testing.T) {
    // you test code
  })
 }
}
```

### Options

You can initialize server with the following options:

* `WithConfigFile()` -- allows to configure server using the JSON configuration file. See [Custom configuration](#custom-configuration) for the file format.
  Mind the order of options when use with `WithConfiguration()`, `WithAddress()` and `WithPort()`.
* `WithConfiguration()` -- allows to configure server with the `Configuration` object.
  Mind the order of options when use with `WithConfigFile()`, `WithAddress()` and `WithPort()`.
* `WithAddress()` -- allows to set up the serving address for the server.
  Mind the order of options when use with `WithConfigFile()` and `WithConfiguration()`.
* `WithPort()` -- allows to set up the port that the server will be listening at.
  Mind the order of options when use with `WithConfigFile()` and `WithConfiguration()`.
* `WithEndpoint()` -- allows to set up the default endpoint path.
  Mind the order of options when use with `WithConfigFile()` and `WithConfiguration()`.
* `WithHandlers()` -- allows to set up the metadata paths and responses when the metadata request is served at the paths.
  Mind the order of options when use with `WithConfigFile()` and `WithConfiguration()`.
* `WithLogger` -- allows to setup a custom `slog.Logger`. If no logger is set up the metadata server writes logs to `io.Discard`.

### Custom configuration

You can define custom configurations using JSON configuration file instead of setting them up in the code.
See [example configurations](examples/) in the repo.
You also can use `WithConfiguration()` option to define the configuration in the code instead of using other `With*` functions.

The JSON file schema is described below:

| Name | Type | Description |
|---|---|---|
| `address` | `string` | IP address of where the server serves the requests. Default value `169.254.169.254`. |
| `port` | `numeric` | Port number at which the server listens. Default value `80`. |
| `endpoint` | `string` | The default path. Together with `address` and `port` it defined the default endpoint and also is used as a prefix for other handler's paths. Sending request to the default endpoint always returns "ok". Default value `computeMetadata/v1`. |
| `shutdownTimeout` | `numeric` | The time in seconds that takes to server to timeout at shutdown. Default value `5` (sec). |
| `metadata` | map | Collection of key-values describing the returned metadata. See next paragraph for more information. |

#### Metadata keys and values

Metadata maps keys to values allowing customization of data that the server returns on different paths. The path is composed of concatinating the `endpoint` with the metadata's key string.
For example, for the default endpoint and the key "project/project-id", the server will respond at the path "/computeMetadata/v1/project/project-id" with the value defined in the metadata map.

Metadata map supports two types of values:

* Static values -- literals that are returned when a request is send using the path of the endpoint + key. Use the following JSON to define the static value:

  ```json
  {
    "value": "STATIC_VALUE"
  }
  ```

* Environment-based values -- the returned value is retrieved from the environment variable which name is defined in the metadata's value.
  When a request is send using the path of the endpoint + key, the server will read and return the value of the environment variable. Use the following JSON to define the environment-based value:

  ```json
  {
    "env": "ENV_VARIABLE_NAME"
  }
  ```

The following example of the custom configuration sets up the server to serve three metadata values at the following paths:

* `/custom/endpoint/static` path will return `always the same`
* `/custom/endpoint/environment_a` path will return the value of the environment variable with the name `X_A`
* `/custom/endpoint/another/static` path will return `this is another always the same value`

```json
{
    "endpoint": "custom/endpoint",
    "metadata": {
        "static": {
            "value": "always the same"
        },
        "environment_a": {
            "env": "X_A"
        },
        "another/static": {
            "value": "this is another always the same value"
        }
    }
}
```

> [!NOTE]
> Configuration values that were not customized keep their default values.
> If no metadata is configured, the server will respond at the path defined by the endpoint only.

### Metadata server IP address

The package does not implement any networking configuration on the local host.
If you code uses a hostname (e.g. `metadata.google.internal` for Google's metadata server), this hostname has to be explicitly configured in your debug environment or your code has to run the metadataserver using IP address instead.
Majority of providers run their metadata servers using `169.254.169.254` IP address.
It is a [link-local address](https://en.wikipedia.org/wiki/Link-local_address).
This means that these addresses are usually have to be configured in the environment prior to use.

Use the following Linux shell command to configure link this address to your localhost interface:

```shell
sudo ip addr add 169.254.169.254/16 dev lo
```

Use the following PowerShell to do the same on Windows:

```powershell
New-NetIPAddress -InterfaceAlias "Loopback" -IPAddress "169.254.169.254" -PrefixLength 16
```

> [!NOTE]
> It is highly unlikely that you already have a link for "169.254.169.254" in your environment.
> However, take precautions not to override the already existing configuration.
