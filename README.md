# Cloud Foundry buildpack: ASP.NET 5

A Cloud Foundry buildpack for ASP.NET 5 applications (tested with [beta6][] applications). For more information about ASP.NET 5 see:

* https://github.com/aspnet/home
* http://docs.asp.net/en/latest/conceptual-overview/aspnet.html

## Usage

```bash
cf push my_app -b https://github.com/cloudfoundry-community/asp.net5-buildpack.git
```

This buildpack will be used if there is a `NuGet.Config` file in the root directory of your project and one or more 'project.json' files anywhere.

Use a 'global.json' file to specify the desired DNX version if different than the latest stable beta release. Also make sure the application includes a 'kestrel' command and the corresponding dependency as that is what the buildpack will use to run the application.

For a basic example see this [Hello World sample][].

## Disconnected environments
Disconnected environments are not fully supported. The .NET Version Manager, .NET Execution Environment and NuGet packages required by the application are always downloaded during staging. The Mono and libuv binaries, however, can be cached or uncached.

## Building

1. Make sure you have fetched submodules

  ```bash
  git submodule update --init
  ```

2. Get latest buildpack dependencies

  ```shell
  BUNDLE_GEMFILE=cf.Gemfile bundle
  ```

3. Build the binary dependencies (optional)

 If you need to rebuild these, to change a version for example, see the included Dockerfiles. They contain comments specifying the commands to run. Then update manifest.yml to point to your files.

4. Build the buildpack

    `uncached` means a TAR files containing Mono and libuv will be downloaded the first time an application is staged, and `cached` means they will be packaged in the buildpack ZIP.

  ```shell
  BUNDLE_GEMFILE=cf.Gemfile bundle exec buildpack-packager [ uncached | cached ]
  ```

5. Use in Cloud Foundry

    Upload the buildpack to your Cloud Foundry and optionally specify it by name
        
    ```bash
    cf create-buildpack custom_aspnet5_buildpack aspnet5_buildpack-cached-custom.zip 1
    cf push my_app -b custom_aspnet5_buildpack
    ```  

## Contributing

Find our guidelines [here](./CONTRIBUTING.md).

## Reporting Issues

Open an issue on this project.


[Hello World sample]: https://github.com/IBM-Bluemix/asp.net5-helloworld
[beta6]: https://github.com/aspnet/Home/releases/tag/v1.0.0-beta6
