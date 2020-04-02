# Cloud Foundry buildpack: .NET Core

[![CF Slack](https://www.google.com/s2/favicons?domain=www.slack.com) Join us on Slack](https://cloudfoundry.slack.com/messages/buildpacks/)

A Cloud Foundry [buildpack](http://docs.cloudfoundry.org/buildpacks/) for .NET Core applications.

For more information about ASP.NET Core see:

* [ASP.NET Github](https://github.com/aspnet/home)
* [Introduction to ASP.NET Core](http://docs.asp.net/en/latest/conceptual-overview/aspnet.html)

### Buildpack User Documentation

Official buildpack documentation can be found at <http://docs.cloudfoundry.org/buildpacks/dotnet-core/index.html>.

### Building the Buildpack

To build this buildpack, run the following commands from the buildpack's directory:

1. Source the .envrc file in the buildpack directory.

   ```bash
   source .envrc
   ```
   To simplify the process in the future, install [direnv](https://direnv.net/) which will automatically source .envrc when you change directories.

1. Install buildpack-packager

    ```bash
    go install github.com/cloudfoundry/libbuildpack/packager/buildpack-packager
    ```

1. Build the buildpack

    ```bash
    buildpack-packager build [ --cached ] [ --stack <stack> ]
    ```

1. Use in Cloud Foundry

   Upload the buildpack to your Cloud Foundry and optionally specify it by name

    ```bash
    cf create-buildpack [BUILDPACK_NAME] [BUILDPACK_ZIP_FILE_PATH] 1
    cf push my_app [-b BUILDPACK_NAME]
    ```

### Testing

Buildpacks use the [Cutlass](https://github.com/cloudfoundry/libbuildpack/tree/master/cutlass) framework for running integration tests against Cloud Foundry. Before running the integration tests, you need to login to your Cloud Foundry using the [cf cli](https://github.com/cloudfoundry/cli):

 ```bash
 cf login -a https://api.your-cf.com -u name@example.com -p pa55woRD
 ```

Note that your user requires permissions to run `cf create-buildpack` and `cf update-buildpack`. To run the integration tests, run the following command from the buildpack's directory:

1. Source the .envrc file in the buildpack directory.

   ```bash
   source .envrc
   ```
   To simplify the process in the future, install [direnv](https://direnv.net/) which will automatically source .envrc when you change directories.

1. Run unit tests

    ```bash
    ./scripts/unit.sh
    ```

1. Run integration tests

    ```bash
    ./scripts/integration.sh
    ```

### Contributing

Find our guidelines [here](./CONTRIBUTING.md).

### Help and Support

Join the #buildpacks channel in our [Slack community](http://slack.cloudfoundry.org/) if you need any further assistance.

## Contributing

Find our guidelines [here](./CONTRIBUTING.md).

### Reporting Issues

Please fill out the issue template fully if you'd like to start an issue for the buildpack.

### Links

* [Hello World sample](https://github.com/IBM-Bluemix/aspnet-core-helloworld)
* [ASP.NET Core 1.0.1](https://github.com/aspnet/Home/releases/tag/1.0.1)
* [Kestrel](https://github.com/aspnet/KestrelHttpServer)
* [.NET Core Application Deployment](https://docs.microsoft.com/en-us/dotnet/articles/core/deploying/index)
* [Specify the path](https://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#-procedure)

