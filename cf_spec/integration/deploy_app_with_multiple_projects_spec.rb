$LOAD_PATH << 'cf_spec'
require 'spec_helper'

describe 'CF ASP.NET Core Buildpack' do
  subject(:app) { Machete.deploy_app(app_name) }
  let(:browser) { Machete::Browser.new(app) }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  context 'deploying an app with multiple projects' do
    let(:app_name) { 'app_with_multiple_projects' }

    it 'compiles both apps' do
      expect(app).to be_running
      expect(app).to have_logged(/Compiling console_app/)
      expect(app).to have_logged(/Compiling asp_web_app/)

      browser.visit_path('/')
      expect(browser).to have_body("Hello, I'm a string!")
      expect(app).to have_logged("Hello from a secondary project!")
    end
  end

end
