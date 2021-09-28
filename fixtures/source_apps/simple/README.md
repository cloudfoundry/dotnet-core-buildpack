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
It is used to test the simple/default .NET Core test case.
