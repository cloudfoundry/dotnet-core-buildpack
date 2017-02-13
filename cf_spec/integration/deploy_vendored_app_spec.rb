$LOAD_PATH << 'cf_spec'
require 'spec_helper'

describe 'deploying vendored dotnet apps' do
  subject(:app) { Machete.deploy_app(app_name) }
  let(:browser) { Machete::Browser.new(app) }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  context 'the app is portable', :cached do
    let(:app_name) { 'asp_vendored' }

    it 'displays a simple text homepage' do
      expect(app).to be_running
      expect(app).to have_logged(/ASP.NET Core buildpack is done creating the droplet/)
      expect(app).not_to have_internet_traffic

      browser.visit_path('/')
      expect(browser).to have_body('Hello World!')
    end
  end

  context 'the app is self contained', :cached do
    let(:app_name) { 'self_contained' }

    it 'displays a simple text homepage' do
      expect(app).to be_running
      expect(app).to have_logged(/ASP.NET Core buildpack is done creating the droplet/)
      expect(app).not_to have_internet_traffic

      browser.visit_path('/')
      expect(browser).to have_body('Hello World!')
    end
  end
end
