# Metadata Server

[![Go Report Card](https://goreportcard.com/badge/github.com/minherz/metadataserver)](https://goreportcard.com/report/github.com/minherz/metadataserver)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/minherz/metadataserver)](https://pkg.go.dev/github.com/minherz/metadataserver)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/minherz/metadataserver)
![Repo license](https://badgen.net/badge/license/Apache%202.0/blue)

This is a simulation of the metadata server that run on cloud environments of such providers as Amazon, Google or Azure.
This server is intended to assist in local debugging of applications that are intended for run in cloud environments and make use of the environment metadata.
Alternative is to debug these applications in the cloud environment which isn't always convenient or to customize the application code to mock or substitute calls to metadata server.

The metadata server listens to port `80` on `169.254.169.254`. If the code uses a hostname (e.g. `metadata.google.internal`) this hostname should be manually configured in the local environment.

The default configuration of the metadata server supports the following endpoint:

* `http://169.254.169.254/computeMetadata/v1`

> [!NOTE]
> Currently metadata server does not support HTTPS endpoints

## Use the package

To use the package do the following:

1. Import it in your code:

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
   ms.Start()
   ```

4. Stop the server:

   ```go
   ms.Stop(context.Background())
   ```

### Options

You can initialize server with the following options:

* `WithConfigFile()` -- allows to configure server using the JSON configuration file. See [Custom configuration](#custom-configuration) for the file format.
  Do not use it together with `WithConfiguration()`, `WithAddress()` and `WithPort()`.
* `WithConfiguration()` -- allows to configure server with the `Configuration` object.
  Do not use it together with `WithConfigFile()`, `WithAddress()` and `WithPort()`.
* `WithAddress()` -- allows to set up the serving address for the server.
  Do not use it together with `WithConfigFile()` and `WithConfiguration()`.
* `WithPort()` -- allows to set up the port that the server will be listening at.
  Do not use it together with `WithConfigFile()` and `WithConfiguration()`.
* `WithLogger` -- allows to setup a custom `slog.Logger`. If no logger is set up the metadata server writes logs to `io.Discard`.

### Custom configuration

You can customize the behavior of the metadata server by setting the following parameters:

| Name | Type | Description |
|---|---|---|
| `address` | `string` | IP address of where the server serves the requests. Default value `169.254.169.254`. |
| `port` | `numeric` | Port number at which the server listens. Default value `80`. |
| `endpoint` | `string` | The root path for all handlers. Server responses with "ok" when a request is sent to this path. The path can omit the starting and ending slashes. Default value `computeMetadata/v1`. |
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
