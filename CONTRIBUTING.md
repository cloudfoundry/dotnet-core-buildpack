# Contributing

## Run the unit tests

1. `BUNDLE_GEMFILE=cf.Gemfile bundle install`
1. `BUNDLE_GEMFILE=cf.Gemfile bundle exec rake spec`

## Run the full test suite (unit and integration)

1. `BUNDLE_GEMFILE=cf.Gemfile bundle install`
1. `BUNDLE_GEMFILE=cf.Gemfile buildpack-build`

## Comply with formatting and style

1. `BUNDLE_GEMFILE=cf.Gemfile bundle exec rubocop -a`

## Pull Requests

1. Fork the project
1. Submit a pull request to the `develop` branch

Please include integration tests (in `cf_spec/integration`) and corresponding fixtures (in `cf_spec/fixtures`) where necessary to cover any functionality that is introduced.
