require 'bundler/setup'
require 'machete'
require 'machete/matchers'
require 'rspec/retry'
require_relative '../lib/buildpack.rb'
require_relative '../lib/buildpack/shell.rb'
require_relative '../lib/buildpack/sdk_info.rb'

`mkdir -p log`
Machete.logger = Machete::Logger.new('log/integration.log')

RSpec.configure do |config|
  config.color = true
  config.tty = true
  config.verbose_retry = true
  config.default_retry_count = 3
  config.default_sleep_interval = 5

  config.filter_run_excluding cached: ENV['BUILDPACK_MODE'] == 'uncached'
  config.filter_run_excluding uncached: ENV['BUILDPACK_MODE'] == 'cached'
end
