This app was generated with the .NET Core CLI by running:
```
dotnet new web -o with_dot_in_name
```
An `AssemblyName` was added to the `.csproj` file with a dot in it:
```
<AssemblyName>some_other.name</AssemblyName>
```
It is used to test that apps with "." in their name can still be built correctly.
