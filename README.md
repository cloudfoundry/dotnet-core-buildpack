# Cloud Foundry buildpack: ASP.NET 5

A Cloud Foundry buildpack for ASP.NET 5 web applications. Tested with [beta8][] applications that target .NET Core.

For more information about ASP.NET 5 see:

* https://github.com/aspnet/home
* http://docs.asp.net/en/latest/conceptual-overview/aspnet.html

## Usage

```bash
cf push my_app -b https://github.com/cloudfoundry-community/asp.net5-buildpack.git
```

This buildpack will be used if there are one or more `project.json` files in the pushed application. 

Also make sure the application includes a `kestrel` or a `web` command and the corresponding Microsoft.AspNet.Server.Kestrel dependency because the buildpack will use [Kestrel][] to run the application.

Use a `global.json` file to specify the desired DNX version if different than the latest stable beta release. Use a `NuGet.Config` file to specify non-default package sources.

For a basic example see this [Hello World sample][].

## Disconnected environments
The binaries in `manifest.yml` can be cached with the buildpack. 

Applications can be pushed with their other dependencies after "publishing" the application like `dnu publish` or `dnu publish --runtime ~/.dnx/runtimes/dnx-coreclr-linux-x64.1.0.0-beta7`. Then push from the `bin/output` director.

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
    cf create-buildpack custom_aspnet5_buildpack aspnet5_buildpack-cached-custom.zip 1
    cf push my_app -b custom_aspnet5_buildpack
    ```  

## Contributing

Find our guidelines [here](./CONTRIBUTING.md).

## Reporting Issues

Open an issue on this project.


[Hello World sample]: https://github.com/IBM-Bluemix/asp.net5-helloworld
[beta8]: https://github.com/aspnet/Home/releases/tag/v1.0.0-beta8
[Kestrel]: https://github.com/aspnet/KestrelHttpServer
