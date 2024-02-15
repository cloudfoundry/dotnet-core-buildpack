This app is used to test the functionality to load the legacy openssl provider
when using cflinuxfs4

The app is tested with the addition of the `BP_OPENSSL_ACTIVATE_LEGACY_PROVIDER` environment
variable set to `true`

See ssl-related integration test in `src/dotnetcore/integration/default_test.go`
