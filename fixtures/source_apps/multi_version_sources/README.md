This app was generated with the .NET Core CLI by running:
```
dotnet new web -o simple_3.1_source
```
A exit handler was added to the `Program.cs` file:
```
            AppDomain.CurrentDomain.ProcessExit +=
                 (sender, eventArgs) => {
                     Console.WriteLine("Goodbye, cruel world!");
                 };
```

Then, a templated `global.json` and `buildpack.yml` file were added.
