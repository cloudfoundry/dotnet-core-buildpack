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

describe AspNetCoreBuildpack::DotnetVersion do
  let(:out) { double(:out) }
  let(:dir) { Dir.mktmpdir }
  let(:manifest_file) { File.join(dir, 'manifest.yml') }
  let(:dotnet_versions_file) { File.join(dir, 'dotnet-versions.yml') }

  let(:dotnet_versions_yml) do
    <<~YAML
       ---
       - dotnet: sdk-version-1
         framework: 0.9.99
       - dotnet: sdk-version-2
         framework: 1.0.0
       - dotnet: sdk-version-3
         framework: 1.0.1
       YAML
  end

  let(:manifest_yml) do
    <<~YAML
       ---
       default_versions:
         - name: dotnet
           version: sdk-version-3
       dependencies:
         - name: dotnet
           version: sdk-version-1
         - name: dotnet
           version: sdk-version-2
         - name: dotnet
           version: sdk-version-3
       YAML
  end

  let(:latest_version) { 'sdk-version-3'.freeze }

  before do
    File.write(dotnet_versions_file, dotnet_versions_yml)
    File.write(manifest_file, manifest_yml)
  end

  after do
    FileUtils.rm_rf(dir)
  end

  subject { described_class.new(dir, manifest_file, dotnet_versions_file, out) }

  describe '#version' do
    context 'global.json does not exist' do
      context '*.runtimeconfig.json does not exist' do
        it 'resolves to the latest version' do
          expect(subject.version).to eq(latest_version)
        end
      end

      context 'invalid *.runtimeconfig.json exists' do
        before do
          json = '{ "runtimeOptions": { "framework": { "name": "Microsoft.NETCore.App" } } }'
          IO.write(File.join(dir, 'testapp.runtimeconfig.json'), json)
        end

        it 'resolves to the latest version' do
          expect(subject.version).to eq(latest_version)
        end
      end

      context '*.runtimeconfig.json has non-included version' do
        before do
          json = '{ "runtimeOptions": { "framework": { "name": "Microsoft.NETCore.App", "version": "99.99.99" } } }'
          IO.write(File.join(dir, 'testapp.runtimeconfig.json'), json)
        end

        it 'resolves to the latest version' do
          expect(subject.version).to eq(latest_version)
        end
      end

      context 'valid *.runtimeconfig.json exists' do
        before do
          json = '{ "runtimeOptions": { "framework": { "name": "Microsoft.NETCore.App", "version": "1.0.0" } } }'
          IO.write(File.join(dir, 'testapp.runtimeconfig.json'), json)
        end

        it 'maps the version specified in *.runtimeconfig.json' do
          expect(subject.version).to eq('sdk-version-2')
        end
      end
    end

    context 'global.json exists' do
      before do
        json = '{ "sdk": { "version": "1.0.0-beta1" } }'
        IO.write(File.join(dir, 'global.json'), json)
      end

      it 'resolves to the specified version' do
        expect(subject.version).to eq('1.0.0-beta1')
      end
    end

    context 'global.json exists with a BOM from Visual Studio in it' do
      before do
        json = "\uFEFF{ \"sdk\": { \"version\": \"1.0.0-beta1\" } }"
        IO.write(File.join(dir, 'global.json'), json)
      end

      it 'resolves to the specified version' do
        expect(subject.version).to eq('1.0.0-beta1')
      end
    end

    context 'invalid global.json exists' do
      before do
        json = '"version": "1.0.0-beta1"'
        IO.write(File.join(dir, 'global.json'), json)
      end

      it 'warns and resolves to the latest version' do
        expect(out).to receive(:warn).with("File #{dir}/global.json is not valid JSON")
        expect(subject.version).to eq(latest_version)
      end
    end

    context 'global.json exists but does not include a version' do
      before do
        json = '{ "projects": [ "src", "test" ] }'
        IO.write(File.join(dir, 'global.json'), json)
      end

      it 'resolves to the latest version' do
        expect(subject.version).to eq(latest_version)
      end
    end
  end
end
