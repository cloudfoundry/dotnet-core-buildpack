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
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'Microsoft.NETCore.App', '1.3.3'))
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'Microsoft.NETCore.App', '1.3.4'))
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'Microsoft.NETCore.App', '1.4.6'))
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'Microsoft.NETCore.App', '1.4.5'))
          end

          it 'returns the latest patch version for each restored major/minor line' do
            expect_any_instance_of(AspNetCoreBuildpack::Out).to receive(:print).with(
              "Detected .NET Core runtime version(s) 1.3.4, 1.4.6 required according to 'dotnet restore'")
            expect(subject.versions).to eq( ['1.3.4', '1.4.6'] )
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

        before do
          FileUtils.mkdir_p(File.join(build_dir, 'prj1'))
          File.write(File.join(build_dir, 'prj1', 'prj1.csproj'), prj1_xml)

          FileUtils.mkdir_p(File.join(build_dir, 'prj2'))
          File.write(File.join(build_dir, 'prj2', 'prj2.csproj'), prj2_xml)
        end

        context '.csproj has RuntimeFrameworkVersion' do
          let(:prj1_xml) do
            <<-XML
<Project Sdk="Microsoft.NET.Sdk.Web">
  <PropertyGroup>
    <TargetFramework>netcoreapp1.0</TargetFramework>
    <DebugType>portable</DebugType>
    <AssemblyName>simple_brats</AssemblyName>
    <OutputType>Exe</OutputType>
    <RuntimeFrameworkVersion>1.2.3</RuntimeFrameworkVersion>
  </PropertyGroup>
</Project>
XML
          end

          let(:prj2_xml) do
            <<-XML
<Project Sdk="Microsoft.NET.Sdk.Web">
  <PropertyGroup>
    <TargetFramework>netcoreapp1.0</TargetFramework>
    <DebugType>portable</DebugType>
    <AssemblyName>simple_brats</AssemblyName>
    <OutputType>Exe</OutputType>
    <RuntimeFrameworkVersion>1.4.5</RuntimeFrameworkVersion>
  </PropertyGroup>
</Project>
XML
          end

          before do
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'microsoft.netcore.app', '1.2.2'))
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'microsoft.netcore.app', '1.2.3'))
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'microsoft.netcore.app', '1.2.4'))
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'microsoft.netcore.app', '1.4.5'))
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'microsoft.netcore.app', '1.4.6'))
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'microsoft.netcore.app', '1.4.7'))
          end

          it 'returns the latest patch version for each restored major/minor line + any specified as RuntimeFramworkVersion in .csproj files' do
            expect_any_instance_of(AspNetCoreBuildpack::Out).to receive(:print).with(
              "Detected .NET Core runtime version(s) 1.2.3, 1.2.4, 1.4.5, 1.4.7 required according to 'dotnet restore'")
            expect(subject.versions).to eq( ['1.2.3', '1.2.4', '1.4.5', '1.4.7'] )
          end
        end

        context '.csproj does not have RuntimeFrameworkVersion' do
          let(:prj1_xml) do
            <<-XML
<Project Sdk="Microsoft.NET.Sdk.Web">
  <PropertyGroup>
    <TargetFramework>netcoreapp1.0</TargetFramework>
    <DebugType>portable</DebugType>
    <AssemblyName>simple_brats</AssemblyName>
    <OutputType>Exe</OutputType>
  </PropertyGroup>
</Project>
XML
          end

          let(:prj2_xml) do
            <<-XML
<Project Sdk="Microsoft.NET.Sdk.Web">
  <PropertyGroup>
    <TargetFramework>netcoreapp1.0</TargetFramework>
    <DebugType>portable</DebugType>
    <AssemblyName>simple_brats</AssemblyName>
    <OutputType>Exe</OutputType>
  </PropertyGroup>
</Project>
XML
          end

          before do
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'microsoft.netcore.app', '1.1.1'))
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'microsoft.netcore.app', '2.2.2'))
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'microsoft.netcore.app', '1.1.2'))
            FileUtils.mkdir_p(File.join(nuget_cache_dir, 'packages', 'microsoft.netcore.app', '2.2.3'))
          end

          it 'returns the latest patch version for each restored major/minor line' do
            expect_any_instance_of(AspNetCoreBuildpack::Out).to receive(:print).with(
              "Detected .NET Core runtime version(s) 1.1.2, 2.2.3 required according to 'dotnet restore'")
            expect(subject.versions).to eq( ['1.1.2', '2.2.3'] )
          end
        end

        context 'dotnet restore detected no framework versions' do
          let(:prj1_xml) { 'does not matter'}
          let(:prj2_xml) { 'does not matter'}

          it 'throws an exception with a helpful message' do
            expect { subject.versions }.to raise_error(RuntimeError, "Unable to determine .NET Core runtime version(s) to install")
          end
        end
      end
    end
  end
end














