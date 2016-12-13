$LOAD_PATH << 'cf_spec'
require 'spec_helper'
require 'rspec'
require 'tmpdir'
require 'fileutils'

describe AspNetCoreBuildpack::SdkInfo do
  class StubUsingSdkInfo
    include AspNetCoreBuildpack::SdkInfo
  end

  subject do
    StubUsingSdkInfo.new
  end

  let(:build_dir)     { Dir.mktmpdir }
  let(:buildpack_dir) { Dir.mktmpdir }
  let(:dotnet_sdk_tools_file) { File.join(buildpack_dir, 'dotnet-sdk-tools.yml') }
  let(:dotnet_sdk_tools_yml) do
    <<-YAML
---
project_json:
- sdk-version-1
- sdk-version-2
msbuild:
- sdk-version-3
- sdk-version-4
  YAML
  end

  let(:sdk_version)      { 'override' }
  let(:sdk_version_file) { File.join(build_dir, '.dotnet', 'VERSION') }

  before do
    FileUtils.mkdir_p(File.join(build_dir, '.dotnet'))
    File.write(sdk_version_file, sdk_version)

    File.write(dotnet_sdk_tools_file, dotnet_sdk_tools_yml)
    allow(subject).to receive(:dotnet_sdk_tools_file).and_return(dotnet_sdk_tools_file)
  end

  after do
    FileUtils.rm_rf(build_dir)
    FileUtils.rm_rf(buildpack_dir)
  end

  describe '#installed_sdk_version' do
    let(:sdk_version) { 'sdk-version' }

    context 'dotnet sdk has been installed' do
      it "returns the sdk version written to .dotnet/VERSION" do
        expect(subject.installed_sdk_version(build_dir)).to eq('sdk-version')
      end
    end

    context 'dotnet sdk has not been installed' do
      before { FileUtils.rm_rf(sdk_version_file) }

      it "errors and states the VERSION file does not exist" do
        expect{ subject.installed_sdk_version(build_dir) }.to raise_error(RuntimeError,
          ".NET SDK version file: #{build_dir}/.dotnet/VERSION does not exist")
      end
    end

  end

  describe '#msbuild?' do
    context 'sdk version uses msbuild' do
      let(:sdk_version) { 'sdk-version-3' }

      it "returns true" do
        expect(subject.msbuild?(build_dir)).to be_truthy
      end
    end

    context 'sdk version uses project.json' do
      let(:sdk_version) { 'sdk-version-2' }

      it "returns false" do
        expect(subject.msbuild?(build_dir)).to be_falsey
      end
    end
  end

  describe '#project_json?' do
    context 'sdk version uses msbuild' do
      let(:sdk_version) { 'sdk-version-4' }

      it "returns true" do
        expect(subject.project_json?(build_dir)).to be_falsey
      end
    end

    context 'sdk version uses project.json' do
      let(:sdk_version) { 'sdk-version-1' }

      it "returns false" do
        expect(subject.project_json?(build_dir)).to be_truthy
      end
    end

  end
end
