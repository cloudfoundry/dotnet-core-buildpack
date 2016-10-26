$LOAD_PATH << 'cf_spec'
require 'spec_helper'

describe 'Deploying an app that uses Nancy framework with Kestrel' do
  let(:app_name) { 'nancy_kestrel_web_app' }
  let(:app) { Machete.deploy_app(app_name) }
  let(:browser) { Machete::Browser.new(app) }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  it 'displays a page served through nancy' do
    expect(app).to be_running
    expect(app).to have_logged(/ASP.NET Core buildpack is done creating the droplet/)

    browser.visit_path('/')
    expect(browser).to have_body('Hello from Nancy running on CoreCLR')
  end
end
