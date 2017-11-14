# Encoding: utf-8
# ASP.NET Core Buildpack
# Copyright 2016 the original author or authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

$LOAD_PATH << 'cf_spec'
require 'spec_helper'
require 'rspec'

describe AspNetCoreBuildpack::DotnetFramework do
  let(:build_dir)          { Dir.mktmpdir }
  let(:deps_dir)          { Dir.mktmpdir }
  let(:deps_idx)          { '55' }
  let(:nuget_cache_dir)    { Dir.mktmpdir }
  let(:dotnet_install_dir) { Dir.mktmpdir }
  let(:shell)              { double(:shell, env: {}) }
  let(:versions)           { %w(4.4.4 5.5.5)}

  let(:out) { double(:out) }
  let(:self_contained_app_dir) { double(:self_contained_app_dir, published_project: 'project1') }
  let(:app_dir) { double(:app_dir, published_project: false, with_project_json: %w(project1 project2)) }

  before do
    allow(shell).to receive(:exec).and_return(0)
    allow(AspNetCoreBuildpack::DotnetFrameworkVersion).to receive(:new).with(any_args).and_return(double(versions: versions ))
  end

  after do
    FileUtils.rm_rf(build_dir)
    FileUtils.rm_rf(nuget_cache_dir)
    FileUtils.rm_rf(dotnet_install_dir)
  end

  subject(:installer) { described_class.new(build_dir, nuget_cache_dir, deps_dir, deps_idx, dotnet_install_dir, shell) }

  describe '#install' do
    context 'both required versions not installed' do
      it 'downloads both frameworks with with compile-extensions' do
        expect(out).to receive(:print).with("Downloading and installing .NET Core runtime 4.4.4")
        expect(shell).to receive(:exec).with(/download_dependency dotnet-framework.4.4.4.linux-amd64.tar.xz \/tmp/, out)
        expect(shell).to receive(:exec).with("mkdir -p #{dotnet_install_dir}; tar xf /tmp/dotnet-framework.4.4.4.linux-amd64.tar.xz -C #{dotnet_install_dir}", out)

        expect(out).to receive(:print).with("Downloading and installing .NET Core runtime 5.5.5")
        expect(shell).to receive(:exec).with(/download_dependency dotnet-framework.5.5.5.linux-amd64.tar.xz \/tmp/, out)
        expect(shell).to receive(:exec).with("mkdir -p #{dotnet_install_dir}; tar xf /tmp/dotnet-framework.5.5.5.linux-amd64.tar.xz -C #{dotnet_install_dir}", out)

        subject.install(out)
      end
    end

    context 'one required version is not installed' do
      before do
        FileUtils.mkdir_p(File.join(dotnet_install_dir, 'shared', 'Microsoft.NETCore.App', '4.4.4'))
      end

      it 'only downloads the framework that has not been installed' do
        expect(out).to receive(:print).with(".NET Core runtime 4.4.4 already installed")

        expect(out).to receive(:print).with("Downloading and installing .NET Core runtime 5.5.5")
        expect(shell).to receive(:exec).with(/download_dependency dotnet-framework.5.5.5.linux-amd64.tar.xz \/tmp/, out)
        expect(shell).to receive(:exec).with("mkdir -p #{dotnet_install_dir}; tar xf /tmp/dotnet-framework.5.5.5.linux-amd64.tar.xz -C #{dotnet_install_dir}", out)

        subject.install(out)
      end
    end

    context 'a version is installed that is not required' do
      let(:versions) { %w(5.5.5) }

      it 'installs the required version' do
        expect(out).to receive(:print).with("Downloading and installing .NET Core runtime 5.5.5")
        expect(shell).to receive(:exec).with(/download_dependency dotnet-framework.5.5.5.linux-amd64.tar.xz \/tmp/, out)
        expect(shell).to receive(:exec).with("mkdir -p #{dotnet_install_dir}; tar xf /tmp/dotnet-framework.5.5.5.linux-amd64.tar.xz -C #{dotnet_install_dir}", out)

        subject.install(out)
      end
    end
  end
end
