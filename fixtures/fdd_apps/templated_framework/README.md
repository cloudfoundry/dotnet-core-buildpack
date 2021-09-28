This app was generated with the .NET Core CLI by running:
```
dotnet new web -o templated_framework
```

Then `dotnet publish --configuration Release --runtime ubuntu.18.04-x64
--self-contained false` was run on the app. The executable named `templated_framework` was
removed, along with other files from the source app.

In the `runtimeconfig.json` file, the `version` field was templated:
```
"version": "<%= framework_version %>"
```
