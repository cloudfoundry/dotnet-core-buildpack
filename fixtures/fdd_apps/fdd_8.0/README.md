# Generating this fixture

This app was generated using `dotnet` CLI v8:
```
dotnet new mvc -o fdd_dotnet_8

dotnet publish fdd_dotnet_8 --configuration Release --runtime linux-x64 --self-contained false --output ./fdd_8.0

rm -rf fde_dotnet_8
```
