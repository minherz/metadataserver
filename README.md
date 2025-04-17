# Metadata Server

This is a simulation of the metadata server that run on cloud environments of such providers as Amazon, Google or Azure.
This server is intended to assist in local debugging of applications that are intended for run in cloud environments and make use of the environment metadata.
Alternative is to debug these applications in the cloud environment which isn't always convenient or to customize the application code to mock or substitute calls to metadata server.

The metadata server listens to port 80 on 169.254.169.254. If the code uses a hostname (e.g. `metadata.google.internal`) this hostname should be manually configured in the local environment.

The default configuration of the metadata server supports the following endpoint:

* `http://metadata.google.internal/computeMetadata/v1`

> [!NOTE!]
> Currently metadata server does not support HTTPS endpoints

## Launch instructions

TBD: download instructions

To launch the metadata server with the default configuration run:

```shell
./metadataserver
```

To run the metadata server with custom configuration run:

```shell
./metadataserver --file="path-to-configuration-file"
```

For more options, run:

```shell
./metadataserver --help
```

## Custom configuration

TBD
