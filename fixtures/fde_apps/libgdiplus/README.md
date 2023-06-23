This app was generated with the .NET Core CLI by running:

```shell
dotnet new web -o libgdiplus
```

Then the following command was run on the app:

```shell
dotnet publish --configuration Release --runtime ubuntu.18.04-x64
--self-contained false libgdiplus
```

Finally, the only files that needed to be copied to the fixture directory are located in
the `bin/Release/Release/net6.0/ubuntu.18.04-x64/publish` directory.