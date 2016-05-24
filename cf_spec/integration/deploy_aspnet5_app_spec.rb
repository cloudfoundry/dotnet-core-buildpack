$LOAD_PATH << 'cf_spec'
require 'spec_helper'

describe 'CF Asp.Net5 Buildpack' do
  subject(:app) { Machete.deploy_app(app_name) }
  let(:browser) { Machete::Browser.new(app) }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  context 'deploying simple web app with internet' do
    let(:app_name) { 'asp_web_app' }

    it 'displays a simple text homepage' do
      expect(app).to be_running
      expect(app).to have_logged(/ASP.NET Core buildpack is done creating the droplet/)

      browser.visit_path('/')
      expect(browser).to have_body('Hello World!')
    end
  end

  context 'deploying an mvc app' do
    let(:app_name) { 'asp_mvc_app' }

    it 'displays a page with an image routed through a controller' do
      expect(app).to be_running
      expect(app).to have_logged(/ASP.NET Core buildpack is done creating the droplet/)

      browser.visit_path('/')
      expect(browser).to have_body("<img src=\"images/ASP-NET-Banners-01.png\" />")
      expect(browser).to have_body('Hello World!')
    end
  end

  context 'deploying an mvc api app' do
    let(:app_name) { 'asp_mvc_api_app' }

    it 'responds to API get requests with json' do
      expect(app).to be_running
      expect(app).to have_logged(/ASP.NET Core buildpack is done creating the droplet/)

      browser.visit_path('/api/products')
      expected_json_response = [
        { Id: 1, Name: 'Computer' },
        { Id: 2, Name: 'Radio' },
        { Id: 3, Name: 'Apple' }
      ]
      expect(browser).to have_body(expected_json_response.to_json)
      expect(browser).to have_header('application/json; charset=utf-8')
    end
  end
end
