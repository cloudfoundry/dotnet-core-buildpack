$: << 'cf_spec'
require 'spec_helper'

describe 'CF Asp.Net5 Buildpack' do
  subject(:app) { Machete.deploy_app(app_name) }
  let(:browser) { Machete::Browser.new(app) }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  context 'deploy static site application with internet' do
    let(:app_name) { 'static_file_internet' }

    it 'responds to http' do
      expect(app).to be_running
      expect(app).to have_logged /ASP.NET 5 buildpack is done creating the droplet/

      browser.visit_path('/')
      expect(browser).to have_body('ASP.NET')
      expect(browser).to have_body('Starter Application')
    end
  end

  context 'deploy project.json application' do
    let(:app_name) { 'mvc_6_application' }

    it 'responds to http' do
      expect(app).to be_running
      expect(app).to have_logged /ASP.NET 5 buildpack is done creating the droplet/

      browser.visit_path('/')
      expect(browser).to have_body("Hi, I'm Nora!")

      browser.visit_path('/env')
      expect(browser).to have_body('Starter Application')
    end
  end
end
