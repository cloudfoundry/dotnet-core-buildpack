# Cloud Foundry buildpack: .NET Core

A Cloud Foundry buildpack for .NET Core applications. Tested with [ASP.NET Core RC2][] applications that target .NET Core.

For more information about ASP.NET Core see:

* https://github.com/aspnet/home
* http://docs.asp.net/en/latest/conceptual-overview/aspnet.html

## Usage

```bash
cf push my_app -b https://github.com/cloudfoundry-community/dotnet-core-buildpack.git
```

This buildpack will be used if there are one or more `project.json` files in the pushed application, or if the application is pushed from the output directory of the `dotnet publish` command. 

Use a `global.json` file to specify the desired .Net CLI version if different than the latest stable beta release.  Use a `NuGet.Config` file to specify non-default package sources.

For a basic example see this [Hello World sample][].

## Legacy DNX support (used for RC1 apps)

With the introduction of support for the Dotnet CLI in buildpack version 0.8, apps which relied on the older DNX toolchain will no longer work with the current buildpack.  If you need to keep your app running on DNX for now until you can update it to use the Dotnet CLI, use the following command:

```bash
cf push my_app -b https://github.com/cloudfoundry-community/dotnet-core-buildpack.git#dnx
```

Keep in mind that this support is provided only to allow users to take some time to update their applications to use the Dotnet CLI, and you should switch to using the main branch of the buildpack (using the command further above) as soon as possible.

## Using samples from the cli-samples repository

The samples provided in the [cli-samples repo](https://github.com/aspnet/cli-samples/) will work with this buildpack but they need a slight modification to the `Main` method.  Before the line `var host = new WebHostBuilder()` add these lines:

```c#
var config = new ConfigurationBuilder()
    .AddCommandLine(args)
    .Build();
```

And then add this line after:
`.UseConfiguration(config)`

You'll also need to add a dependency to project.json:
`"Microsoft.Extensions.Configuration.CommandLine": "1.0.0-rc2-final",`

And a using statement to the file which contains your `Main` method:
`using Microsoft.Extensions.Configuration;`

Example `Main` method:

```c#
public static void Main(string[] args)
{
    var config = new ConfigurationBuilder()
        .AddCommandLine(args)
        .Build();

    var host = new WebHostBuilder()
        .UseKestrel()
        .UseConfiguration(config)
        .UseStartup<Startup>()
        .Build();
    host.Run();
}
```

## Deploying apps with multiple projects

To deploy an app which contains multiple projects, you will need to specify which project you want the buildpack to run as the main project.  This can be done by creating a `.deployment` file in the root folder of the solution which sets the path to the main project.  The path to the main project can be specified as the project folder or the project file (.xproj or .csproj).

For a solution which contains three projects (MyApp.DAL, MyApp.Services, and MyApp.Web which are contained in the "src" folder) where MyApp.Web is the main project to run, the format of the `.deployment` file would be as follows:

```text
[config]
project = src/MyApp.Web
```

In this case, the buildpack would automatically compile the MyApp.DAL and MyApp.Services projects if they are listed as dependencies in the main project's (MyApp.Web) `project.json` file, but the buildpack would only attempt to execute the main project with `dotnet run -p src/MyApp.Web`.  The path to MyApp.Web could also be specified as `project = src/MyApp.Web/MyApp.Web.xproj` (assuming this project is an xproj project).

## Disconnected environments

The binaries in `manifest.yml` can be cached with the buildpack.

Applications can be pushed with their other dependencies after "publishing" the application like `dotnet publish -r ubuntu.14.04-x64`.  Then push from the `bin/<Debug|Release>/<framework>/<runtime>/publish` directory.

For this publish command to work, you will need to make some changes to your application code to ensure that the dotnet cli publishes it as a self-contained application rather than a portable application.

See [Types of portability in .Net Core][] for more information on how to make the required changes to publish your application as a self-contained application.

Also note that if you are using a `manifest.yml` file in your application, you can [specify the path][] in your manifest.yml to point to the publish output folder so that you don't have to be in that folder to push the application to Cloud Foundry.

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

## Unit Testing


Having performed the steps from Building:

  ```shell
  BUNDLE_GEMFILE=cf.Gemfile bundle exec rspec
  ```

### Integration Testing

Integration tests are run using [Machete](https://github.com/cloudfoundry/machete).

To run the tests:

```
CF_PASSWORD=admin BUNDLE_GEMFILE=cf.Gemfile bundle exec buildpack-build --host=local.pcfdev.io
```


## Contributing

Find our guidelines [here](./CONTRIBUTING.md).

## Reporting Issues

Open an issue on this project.


[Hello World sample]: https://github.com/IBM-Bluemix/aspnet-core-helloworld
[ASP.NET Core RC2]: https://github.com/aspnet/Home/releases/tag/1.0.0-rc2-final
[Kestrel]: https://github.com/aspnet/KestrelHttpServer
[Types of portability in .Net Core]: http://dotnet.github.io/docs/core-concepts/app-types.html
[specify the path]: http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#path
