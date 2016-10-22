$LOAD_PATH << 'cf_spec'
require 'spec_helper'
require 'capybara/poltergeist'
require 'capybara/rspec'
require 'phantomjs'

describe 'Deploying an app that relies on Node libraries during staging', type: :feature do
  subject(:app) { Machete.deploy_app(app_name) }
  let(:browser) { Machete::Browser.new(app) }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  context 'deploying an app using angular' do
    let(:app_name) { 'app_using_angular' }

    before do
      Capybara.register_driver :poltergeist do |app|
        Capybara::Poltergeist::Driver.new(app, phantomjs: Phantomjs.path)
      end

      Capybara.current_driver = :poltergeist
      Capybara.run_server = false

      minimum_acceptable_cf_api_version = '2.57.0'
      skip_reason = ".profile script functionality not supported before CF API version #{minimum_acceptable_cf_api_version}"
      Machete::RSpecHelpers.skip_if_cf_api_below(version: minimum_acceptable_cf_api_version, reason: skip_reason)
    end

    it 'displays a javascript homepage' do
      expect(app).to be_running
      expect(app).to have_logged(/ASP.NET Core buildpack is done creating the droplet/)
      visit "http://#{browser.base_url}"
      expect(page).to have_content 'My First Angular 2 App'
    end
  end
end
