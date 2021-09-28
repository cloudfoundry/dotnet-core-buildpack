This app was generated with the .NET Core CLI version 3.1.413 by running:
```
dotnet new mvc -o templated_runtime
```
The `RuntimeFrameworkVersion` was then added and templated in the `.csproj`
file:

```
<RuntimeFrameworkVersion><%= runtime_version %></RuntimeFrameworkVersion>
```
