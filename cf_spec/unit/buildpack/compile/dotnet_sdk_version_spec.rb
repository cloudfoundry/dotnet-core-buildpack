# Encoding: utf-8
# ASP.NET Core Buildpack
# Copyright 2014-2016 the original author or authors.
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
require 'tmpdir'
require 'fileutils'

describe AspNetCoreBuildpack::DotnetSdkVersion do
  let(:dir)               { Dir.mktmpdir }
  let(:deps_dir)               { Dir.mktmpdir }
  let(:deps_idx)               {'88'}
  let(:manifest_file)     { File.join(dir, 'manifest.yml') }
  let(:dotnet_tools_file) { File.join(dir, 'dotnet-sdk-tools.yml') }
  let(:deprecation_warning) do
    "Support for project.json in the .NET Core buildpack will\n" +
    "be deprecated. For more information see:\n" +
    "https://blogs.msdn.microsoft.com/dotnet/2016/11/16/announcing-net-core-tools-msbuild-alpha"
  end

  let(:manifest_yml) do
    <<-YAML
---
default_versions:
- name: dotnet
  version: sdk-version-4
dependencies:
- name: dotnet
  version: sdk-version-1
- name: dotnet
  version: sdk-version-2
- name: dotnet
  version: sdk-version-3
- name: dotnet
  version: sdk-version-4
  YAML
  end

  let(:dotnet_tools_yml) do
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

  let(:default_version) { 'sdk-version-2'.freeze }

  subject { described_class.new(dir, deps_dir, deps_idx, manifest_file) }

  before do
    File.write(manifest_file, manifest_yml)
    File.write(dotnet_tools_file, dotnet_tools_yml)
  end

  after do
    FileUtils.rm_rf(dir)
  end

  describe '#version' do
    context 'global.json does not exist' do
      context 'a project.json file exists and no *.csproj file exists' do
        before do
          FileUtils.mkdir_p(File.join(dir, 'src', 'app'))
          File.write(File.join(dir, 'src', 'app', 'project.json'), 'xxx')
        end

        it 'picks the default version and warns the user project.json will be deprecated' do
          expect_any_instance_of(AspNetCoreBuildpack::Out).to receive(:warn).with(deprecation_warning)
          expect(subject.version).to eq('sdk-version-2')
        end
      end

      context 'a *.csproj file exists and no project.json file exists' do
        before do
          FileUtils.mkdir_p(File.join(dir, 'src', 'app'))
          File.write(File.join(dir, 'src', 'app', 'app.csproj'), 'xxx')
        end

        it 'picks the latest version of the SDK with msbuild' do
          expect(subject.version).to eq('sdk-version-4')
        end
      end

      context 'a *.csproj file exists and a project.json file exists' do
        let(:warning) do
          "Found both project.json and MSBuild projects in app:\n" +
          "MSBuild projects: src/app/app.csproj\n" +
          "Directories with project.json: src/app\n" +
          "Please provide a global.json file that specifies the\n" +
          'correct .NET SDK version for this app'
        end

        before do
          FileUtils.mkdir_p(File.join(dir, 'src', 'app'))
          File.write(File.join(dir, 'src', 'app', 'app.csproj'), 'xxx')
          File.write(File.join(dir, 'src', 'app', 'project.json'), 'xxx')
        end

        context 'the user does not supply a global.json file to indicate tooling' do
          it 'logs helpful information and throws an error' do
            expect_any_instance_of(AspNetCoreBuildpack::Out).to receive(:warn).with(warning)

            expect { subject.version }.to raise_error(RuntimeError, "App contains both project.json and MSBuild projects")
          end
        end
      end

      context 'a *.fsproj file exists' do
        before do
          FileUtils.mkdir_p(File.join(dir, 'src', 'app'))
          File.write(File.join(dir, 'src', 'app', 'app.fsproj'), 'xxx')
        end

        it 'picks the latest version of the SDK with fsharp support' do
          expect(subject.version).to eq('1.1.0')
        end
      end
    end

    context 'global.json exists' do
      context 'and version exists in dotnet_tools_yml' do
        before do
          json = '{ "sdk": { "version": "sdk-version-3" } }'
          IO.write(File.join(dir, 'global.json'), json)
        end

        it 'resolves to the specified version' do
          expect(subject.version).to eq('sdk-version-3')
        end
      end

      context 'and version is missing in dotnet_tools_yml' do
        before do
          json = '{ "sdk": { "version": "1.0.0-beta1" } }'
          IO.write(File.join(dir, 'global.json'), json)
          expect_any_instance_of(AspNetCoreBuildpack::Out).to receive(:warn).with("SDK 1.0.0-beta1 not available,\nusing the default SDK version(sdk-version-4)")
        end

        it 'resolves to the default version' do
          expect(subject.version).to eq('sdk-version-4')
        end
      end
    end

    context 'global.json exists with a BOM from Visual Studio in it' do
      context 'and version exists in dotnet_tools_yml' do
        before do
          json = "\uFEFF{ \"sdk\": { \"version\": \"sdk-version-3\" } }"
          IO.write(File.join(dir, 'global.json'), json)
        end

        it 'resolves to the specified version' do
          expect(subject.version).to eq('sdk-version-3')
        end
      end

      context 'and version is missing in dotnet_tools_yml' do
        before do
          json = "\uFEFF{ \"sdk\": { \"version\": \"1.0.0-beta1\" } }"
          IO.write(File.join(dir, 'global.json'), json)
          expect_any_instance_of(AspNetCoreBuildpack::Out).to receive(:warn).with("SDK 1.0.0-beta1 not available,\nusing the default SDK version(sdk-version-4)")
        end

        it 'resolves to the default version' do
          expect(subject.version).to eq('sdk-version-4')
        end
      end
    end

    context 'invalid global.json exists' do

      before do
        json = '"version": "1.0.0-beta1"'
        IO.write(File.join(dir, 'global.json'), json)
      end

      context 'a project.json file exists and no *.csproj file exists' do
        before do
          FileUtils.mkdir_p(File.join(dir, 'src', 'app'))
          File.write(File.join(dir, 'src', 'app', 'project.json'), 'xxx')
        end

        it 'it warns and picks the default version' do
          expect_any_instance_of(AspNetCoreBuildpack::Out).to receive(:warn).with("File #{dir}/global.json is not valid JSON")
          expect_any_instance_of(AspNetCoreBuildpack::Out).to receive(:warn).with(deprecation_warning)

          expect(subject.version).to eq('sdk-version-2')
        end
      end

      context 'a *.csproj file exists and no project.json file exists' do
        before do
          FileUtils.mkdir_p(File.join(dir, 'src', 'app'))
          File.write(File.join(dir, 'src', 'app', 'app.csproj'), 'xxx')
        end

        it 'it warns and picks the latest version of the SDK with msbuild' do
          expect_any_instance_of(AspNetCoreBuildpack::Out).to receive(:warn).with("File #{dir}/global.json is not valid JSON")
          expect(subject.version).to eq('sdk-version-4')
        end
      end
    end

    context 'global.json exists but does not include a version' do
      before do
        json = '{ "projects": [ "src", "test" ] }'
        IO.write(File.join(dir, 'global.json'), json)
      end

      context 'a project.json file exists and no *.csproj file exists' do
        before do
          FileUtils.mkdir_p(File.join(dir, 'src', 'app'))
          File.write(File.join(dir, 'src', 'app', 'project.json'), 'xxx')
        end

        it 'it picks the default version' do
          expect_any_instance_of(AspNetCoreBuildpack::Out).to receive(:warn).with(deprecation_warning)
          expect(subject.version).to eq('sdk-version-2')
        end
      end

      context 'a *.csproj file exists and no project.json file exists' do
        before do
          FileUtils.mkdir_p(File.join(dir, 'src', 'app'))
          File.write(File.join(dir, 'src', 'app', 'app.csproj'), 'xxx')
        end

        it 'it picks the latest version of the SDK with msbuild' do
          expect(subject.version).to eq('sdk-version-4')
        end
      end
    end
  end
end
