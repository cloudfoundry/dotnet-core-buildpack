# Cloud Foundry buildpack: .NET Core

A Cloud Foundry buildpack for .NET Core applications.

For more information about ASP.NET Core see:

* [ASP.NET Github](https://github.com/aspnet/home)
* [Introduction to ASP.NET Core](http://docs.asp.net/en/latest/conceptual-overview/aspnet.html)

## Usage

```bash
cf push my_app -b https://github.com/cloudfoundry-community/dotnet-core-buildpack.git
```

## Buildpack User Documentation

Official buildpack documentation can be found at <http://docs.cloudfoundry.org/buildpacks/dotnet-core/index.html>.

## Building the Buildpack

These steps only apply to admins who wish to install the buildpack into their Cloud Foundry deployment. They are meant to be run in a Linux shell and assume that git, Ruby, and the bundler gem are already installed.

1. Make sure you have fetched submodules

  ```bash
  git submodule update --init
  ```

1. Get latest buildpack dependencies

  ```bash
  BUNDLE_GEMFILE=cf.Gemfile bundle
  ```

1. Build the binary dependencies (optional)

If you need to rebuild these, to change a version for example, see the included Dockerfiles. They contain comments specifying the commands to run. Then update manifest.yml to point to your files.

1. Build the buildpack

`uncached` means the buildpack's binary dependencies will be downloaded the first time an application is staged, and `cached` means they will be packaged in the buildpack ZIP.

  ```bash
  BUNDLE_GEMFILE=cf.Gemfile bundle exec buildpack-packager [ --uncached | --cached ]
  ```

1. Use in Cloud Foundry

Upload the buildpack to your Cloud Foundry and optionally specify it by name

  ```bash
  cf create-buildpack custom_dotnet-core_buildpack dotnet-core_buildpack-cached-custom.zip 1
  cf push my_app -b custom_dotnet-core_buildpack
  ```

## Unit Testing

Having performed the steps from Building:

  ```shell
  BUNDLE_GEMFILE=cf.Gemfile bundle exec rake spec
  ```

### Integration Testing

Integration tests are run using [Machete](https://github.com/cloudfoundry/machete).

To run all the tests (unit and integration):

```bash
CF_PASSWORD=admin BUNDLE_GEMFILE=cf.Gemfile bundle exec buildpack-build --host=local.pcfdev.io
```

## Contributing

Find our guidelines [here](./CONTRIBUTING.md).

## Reporting Issues

Open an issue on this project.

## Links

[Hello World sample]: https://github.com/IBM-Bluemix/aspnet-core-helloworld
[ASP.NET Core 1.0.1]: https://github.com/aspnet/Home/releases/tag/1.0.1
[Kestrel]: https://github.com/aspnet/KestrelHttpServer
[.NET Core Application Deployment]: https://docs.microsoft.com/en-us/dotnet/articles/core/deploying/index
[Specify the path]: http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#path
