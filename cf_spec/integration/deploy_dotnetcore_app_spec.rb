$LOAD_PATH << 'cf_spec'
require 'spec_helper'

describe 'CF ASP.NET Core Buildpack' do
  subject(:app) { Machete.deploy_app(app_name) }
  let(:browser) { Machete::Browser.new(app) }

  after do
    Machete::CF::DeleteApp.new.execute(app)
  end

  context 'deploying simple web app with internet' do
    let(:app_name) { 'asp' }

    it 'displays a simple text homepage' do
      expect(app).to be_running
      expect(app).to have_logged(/ASP.NET Core buildpack is done creating the droplet/)

      browser.visit_path('/')
      expect(browser).to have_body('Hello World!')
    end

    context 'with BP_DEBUG enabled' do
      subject(:app) { Machete.deploy_app(app_name, env: { 'BP_DEBUG' => '1' }) }

      it 'logs dotnet run verbose output' do
        expect(app).to be_running
        expect(app).to have_logged(/Process ID:/)
        expect(app).to have_logged(/Running \/home\/vcap\/app\/\.dotnet\/dotnet/)
      end
    end
  end

  context 'deploying an mvc app' do
    let(:app_name) { 'asp_mvc' }

    it 'displays a page served through a controller and view' do
      expect(app).to be_running
      expect(app).to have_logged(/ASP.NET Core buildpack is done creating the droplet/)

      browser.visit_path('/')
      expect(browser).to have_body('Hello! Served via a controller and view!')
    end
  end

  context 'deploying an mvc app with node prerendering' do
    let(:app_name) { 'asp_prerender_node' }

    it 'displays a page rendered by node' do
      expect(app).to be_running
      expect(app).to have_logged(/ASP.NET Core buildpack is done creating the droplet/)

      browser.visit_path('/')
      expect(browser).to have_body('1 + 2 = 3')
    end
  end

  context 'deploying an mvc api app' do
    let(:app_name) { 'asp_mvc_api' }

    it 'responds to API get requests with json' do
      expect(app).to be_running
      expect(app).to have_logged(/ASP.NET Core buildpack is done creating the droplet/)

      browser.visit_path('/api/products')
      expected_json_response = [
        { id: 1, name: 'Computer' },
        { id: 2, name: 'Radio' },
        { id: 3, name: 'Apple' }
      ]
      expect(browser).to have_body(expected_json_response.to_json)
      expect(browser).to have_header('application/json; charset=utf-8')
    end
  end

  context 'deploying simple web app in proxied environment', :uncached do
    let(:app_name) { 'console_app' }

    it 'displays a simple text homepage' do
      expect(app).to have_logged(/Hello World/)
      expect(app).to use_proxy_during_staging
    end
  end

  context 'deploying an msbuild app with RuntimeIdentfier' do
    let(:app_name) { 'self_contained_msbuild' }

    it 'displays a simple text homepage' do
      expect(app).to be_running
      expect(app).to have_logged(%r{Removing /tmp/app/\.dotnet})
      expect(app).to have_logged(%r{started using .* \./msbuild_self_contained })

      browser.visit_path('/')
      expect(browser).to have_body('Hello World!')
    end
  end

  context 'deploying an msbuild app with BundlerMinifier.Core' do
    let(:app_name) { 'with_bundler_minifier' }

    it 'does not install dotnet 1.1.0' do
      expect(app).to be_running
      expect(app).not_to have_logged('Downloading and installing .NET Core runtime 1.1.0')

      browser.visit_path('/')
      expect(browser).to have_body('Hello World!')
    end
  end
end
