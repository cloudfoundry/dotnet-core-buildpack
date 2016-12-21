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

describe AspNetCoreBuildpack::DotnetFrameworkVersion do
  let(:build_dir)             { Dir.mktmpdir }
  let(:nuget_cache_dir)       { Dir.mktmpdir}
  let(:app_uses_msbuild)      { false }
  let(:app_uses_project_json) { false }

  subject { described_class.new(build_dir, nuget_cache_dir) }

  before do
    allow(subject).to receive(:msbuild?).and_return(app_uses_msbuild)
    allow(subject).to receive(:project_json?).and_return(app_uses_project_json)
  end

  after do
    FileUtils.rm_rf(build_dir)
    FileUtils.rm_rf(nuget_cache_dir)
  end

  describe '#versions' do
    context '*.runtimeconfig.json exists' do
      context 'valid *.runtimeconfig.json exists' do
        before do
          json = '{ "runtimeOptions": { "framework": { "name": "Microsoft.NETCore.App", "version": "1.0.0" } } }'
          IO.write(File.join(build_dir, 'testapp.runtimeconfig.json'), json)
        end

        it 'returns the framework version specified in *.runtimeconfig.json' do
          expect_any_instance_of(AspNetCoreBuildpack::Out).to receive(:print).with(
            "Detected .NET Core runtime version 1.0.0 in #{build_dir}/testapp.runtimeconfig.json")
          expect(subject.versions).to eq( ['1.0.0'] )
        end
      end

      context '*.runtimeconfig.json does not contain a framework version' do
        before do
          json = '{ "runtimeOptions": { "framework": { "name": "Microsoft.NETCore.App" } } }'
          IO.write(File.join(build_dir, 'testapp.runtimeconfig.json'), json)
        end

        it 'returns an empty array' do
          expect(subject.versions).to eq ( [] )
        end
      end

      context '*.runtimeconfig.json is invalid json' do
        before do
          json = '{ "runtimeOptions": "badjson"  "framework": { "name": "Microsoft.NETCore.App" } } }'
          IO.write(File.join(build_dir, 'testapp.runtimeconfig.json'), json)
        end

        it 'throws an exception with a helpful message' do
          expect { subject.versions }.to raise_error(RuntimeError, "#{build_dir}/testapp.runtimeconfig.json contains invalid JSON")
        end
      end
    end

    context '*.runtimeconfig.json does not exist' do
      context 'with project.json' do
        let(:app_uses_project_json) { true }

        context 'dotnet restore detected required frameworks' do
          before do
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'Microsoft.NETCore.App', '3.3.3'))
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'Microsoft.NETCore.App', '4.4.4'))
          end

          it 'returns the restored framework versions' do
            expect_any_instance_of(AspNetCoreBuildpack::Out).to receive(:print).with(
              "Detected .NET Core runtime version(s) 3.3.3, 4.4.4 required according to 'dotnet restore'")
            expect(subject.versions).to eq( ['3.3.3', '4.4.4'] )
          end
        end

        context 'dotnet restore detected no framework versions' do
          it 'throws an exception with a helpful message' do
            expect { subject.versions }.to raise_error(RuntimeError, "Unable to determine .NET Core runtime version(s) to install")
          end
        end
      end

      context 'with .csproj' do
        let(:app_uses_msbuild) { true }

        context 'dotnet restore detected required frameworks' do
          before do
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'microsoft.netcore.app', '1.1.1'))
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'microsoft.netcore.app', '2.2.2'))
          end

          it 'returns the restored framework versions' do
            expect_any_instance_of(AspNetCoreBuildpack::Out).to receive(:print).with(
              "Detected .NET Core runtime version(s) 1.1.1, 2.2.2 required according to 'dotnet restore'")
            expect(subject.versions).to eq( ['1.1.1', '2.2.2'] )
          end
        end

        context 'dotnet restore detected no framework versions' do
          it 'throws an exception with a helpful message' do
            expect { subject.versions }.to raise_error(RuntimeError, "Unable to determine .NET Core runtime version(s) to install")
          end
        end
      end
    end
  end
end














