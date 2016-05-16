## v0.8.0 May 16, 2016
- Switch to .NET CLI from DNX
- Add support for RC2 apps
- Remove support for RC1 and lower apps using DNX
- detect now looks for project.json files or *.runtimeconfig.json files from publish output
- global.json is no longer used to specify runtime version
- Remove support for lucid64 stack

## v0.7.0 Sep 30, 2015
- Make NuGet.Config optional, detect only looks for project.json files
- Support published apps for offline mode and faster staging
- Remove Mono
- Switch to .NET Core

## v0.6.0 Sep 10, 2015
- Add support for beta7 apps

## v0.5.0 Aug 17, 2015

- Removed Nowin server, replaced with use of Kestrel which allows the buildpack to run beta4 - beta6 apps.

## v0.4.0 Aug 05, 2015

- Add support for beta6 apps
- Remove support for beta5 apps

## v0.3.0 Jul 20, 2015

- Add support for beta5 apps
- Remove support for beta4 apps
- Fix Mono missing trusted root certificates

## v0.2.0 Jun 22, 2015

- Add support for beta4 apps
- Remove support for beta3 apps
- Update Mono version from 3.12.1 to 4.0.1

## v0.1.0 May 05, 2015

- Initial buildpack to run beta3 apps
