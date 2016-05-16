# Cloud Foundry buildpack: ASP.NET Core

A Cloud Foundry buildpack for ASP.NET Core web applications. Tested with [RC2][] applications that target .NET Core.

For more information about ASP.NET Core see:

* https://github.com/aspnet/home
* http://docs.asp.net/en/latest/conceptual-overview/aspnet.html

## Usage

```bash
cf push my_app -b https://github.com/cloudfoundry-community/asp.net5-buildpack.git
```

This buildpack will be used if there are one or more `project.json` files in the pushed application, or if the application is pushed from the output directory of the `dotnet publish` command. 

Use a `NuGet.Config` file to specify non-default package sources.

For a basic example see this [Hello World sample][].

## Legacy DNX support (used for RC1 apps)

With the introduction of support for the Dotnet CLI in buildpack version 0.8, apps which relied on the older DNX toolchain will no longer work with the current buildpack.  If you need to keep your app running on DNX for now until you can update it to use the Dotnet CLI, use the following command:

```bash
cf push my_app -b https://github.com/cloudfoundry-community/asp.net5-buildpack.git#dnx
```

Keep in mind that this support provided to allow users time to update their apps to use the Dotnet CLI, and you should switch to using the main branch of the buildpack (using the command further above) as soon as possible.

## Disconnected environments

The binaries in `manifest.yml` can be cached with the buildpack. 

Applications can be pushed with their other dependencies after "publishing" the application like `dotnet publish`. Then push from the `bin/<Debug|Release>/<framework>/publish` directory.

## Building

These steps only apply to admins who wish to install the buildpack into their Cloud Foundry deployment. They are meant to be run in a Linux shell and assume that git, Ruby, and the bundler gem are already installed.  

1. Make sure you have fetched submodules

  ```bash
  git submodule update --init
  ```

2. Get latest buildpack dependencies

  ```bash
  BUNDLE_GEMFILE=cf.Gemfile bundle
  ```

3. Build the binary dependencies (optional)

 If you need to rebuild these, to change a version for example, see the included Dockerfiles. They contain comments specifying the commands to run. Then update manifest.yml to point to your files.

4. Build the buildpack

    `uncached` means the buildpack's binary dependencies will be downloaded the first time an application is staged, and `cached` means they will be packaged in the buildpack ZIP.

  ```bash
  BUNDLE_GEMFILE=cf.Gemfile bundle exec buildpack-packager [ uncached | cached ]
  ```

5. Use in Cloud Foundry

    Upload the buildpack to your Cloud Foundry and optionally specify it by name

    ```bash
    cf create-buildpack custom_aspnetcore_buildpack aspnetcore_buildpack-cached-custom.zip 1
    cf push my_app -b custom_aspnetcore_buildpack
    ```

## Contributing

Find our guidelines [here](./CONTRIBUTING.md).

## Reporting Issues

Open an issue on this project.


[Hello World sample]: https://github.com/IBM-Bluemix/asp.net5-helloworld
[RC2]: https://github.com/aspnet/Home/releases/tag/1.0.0-rc2-final
[Kestrel]: https://github.com/aspnet/KestrelHttpServer
