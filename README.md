# Cloud Foundry buildpack: ASP.NET 5

A Cloud Foundry buildpack for ASP.NET 5 ([beta3][]) apps. For more information about ASP.NET 5 see:

* https://github.com/aspnet/home
* http://docs.asp.net/en/latest/conceptual-overview/aspnet.html

## Usage

```bash
cf push my_app -b https://github.com/cloudfoundry-community/asp.net5-buildpack.git
```

This buildpack will be used if there is a `NuGet.Config` file in the root directory of your project.

The application structure would then be:
- NuGet.Config
- src
  - project directory
    - Startup.cs
    - project.json
    - wwwroot
    - Models/Views/Controllers or other directories

For a basic example see this [Hello World sample][].

## Disconnected environments
Disconnected environments are not fully supported. The .NET Version Manager, .NET Execution Environment and NuGet packages required by the application are always downloaded during staging. The Mono binaries, however, can be cached or uncached.

## Building

1. Make sure you have fetched submodules

  ```bash
  git submodule update --init
  ```

2. Get latest buildpack dependencies

  ```shell
  BUNDLE_GEMFILE=cf.Gemfile bundle
  ```

3. Build the buildpack

    `uncached` means a TAR file containing Mono will be downloaded the first time an application is staged, and `cached` means it will be packaged in the buildpack ZIP.

  ```shell
  BUNDLE_GEMFILE=cf.Gemfile bundle exec buildpack-packager [ uncached | cached ]
  ```

4. Use in Cloud Foundry

    Upload the buildpack to your Cloud Foundry and optionally specify it by name
        
    ```bash
    cf create-buildpack custom_aspnet5_buildpack aspnet5_buildpack-cached-custom.zip 1
    cf push my_app -b custom_aspnet5_buildpack
    ```  

## Contributing

Find our guidelines [here](./CONTRIBUTING.md).

## Reporting Issues

Open an issue on this project.


[Hello World sample]: https://github.com/opiethehokie/asp.net5-helloworld
[beta3]: https://github.com/aspnet/Home/releases/tag/v1.0.0-beta3
