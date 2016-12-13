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
require 'yaml'
require 'tmpdir'
require 'fileutils'

describe AspNetCoreBuildpack::Releaser do
  let(:build_dir)           { Dir.mktmpdir }
  let(:cache_dir)           { Dir.mktmpdir }
  let(:out)                 { AspNetCoreBuildpack::Out.new }
  let(:is_sdk_project_json) { 'override' }
  let(:is_sdk_msbuild)      { 'override' }

  let(:manifest_dir)  { Dir.mktmpdir }
  let(:manifest_file) { File.join(manifest_dir, 'manifest.yml') }
  let(:manifest_contents) do
    <<-YAML
doesn't matter for these tests
    YAML
  end

  before do
    File.write(manifest_file, manifest_contents)
    FileUtils.mkdir_p(File.join(build_dir, AspNetCoreBuildpack::DotnetSdkInstaller.new(build_dir, cache_dir, manifest_file, out).cache_dir))
    allow(subject).to receive(:msbuild?).and_return(is_sdk_msbuild)
    allow(subject).to receive(:project_json?).and_return(is_sdk_project_json)
    allow_any_instance_of(AspNetCoreBuildpack::AppDir).to receive(:msbuild?).and_return(is_sdk_msbuild)
    allow_any_instance_of(AspNetCoreBuildpack::AppDir).to receive(:project_json?).and_return(is_sdk_project_json)
  end

  after do
    FileUtils.rm_rf(manifest_dir)
  end

  describe '#release' do
    context 'project.json and *.csproj does not exist in source code project' do
      let(:is_sdk_project_json) { false }
      let(:is_sdk_msbuild)      { false }

      it 'raises an error because dotnet run command will not work' do
        expect { subject.release(build_dir) }.to raise_error(/No project could be identified to run/)
      end
    end

    context 'project.json does not exist in published project' do
      let(:is_sdk_project_json) { false }
      let(:is_sdk_msbuild)      { false }

      let(:web_process) do
        yml = YAML.load(subject.release(build_dir))
        yml.fetch('default_process_types').fetch('web')
      end

      before do
        File.open(File.join(build_dir, 'proj1.runtimeconfig.json'), 'w') { |f| f.write('a') }
      end

      context 'project is self-contained' do
        before do
          File.open(File.join(build_dir, 'proj1'), 'w') { |f| f.write('a') }
        end

        it 'does not raise an error because project.json is not required' do
          expect { subject.release(build_dir) }.not_to raise_error
        end

        it 'runs native binary for the project which has a runtimeconfig.json file' do
          expect(web_process).to match('proj1')
        end
      end

      context 'project is a portable project' do
        before do
          File.open(File.join(build_dir, 'proj1.dll'), 'w') { |f| f.write('a') }
        end

        it 'runs dotnet <dllname> for the project which has a runtimeconfig.json file' do
          expect(web_process).to match('dotnet proj1.dll')
        end
      end
    end

    context 'project.json exists' do
      let(:is_sdk_project_json) { true }
      let(:is_sdk_msbuild)      { false }

      let(:proj1) { File.join(build_dir, 'foo').tap { |f| Dir.mkdir(f) } }
      let(:project_json) { '{"commands": {"kestrel": "whatever"}}' }

      let(:profile_d_script) do
        allow_any_instance_of(AspNetCoreBuildpack::DotnetSdkInstaller).to receive(:cached?).and_return(true)
        allow_any_instance_of(AspNetCoreBuildpack::LibunwindInstaller).to receive(:cached?).and_return(true)
        subject.release(build_dir)
        IO.read(File.join(build_dir, '.profile.d', 'startup.sh'))
      end

      let(:web_process) do
        yml = YAML.load(subject.release(build_dir))
        yml.fetch('default_process_types').fetch('web')
      end

      before do
        File.open(File.join(proj1, 'project.json'), 'w') do |f|
          f.write project_json
        end
      end

      it 'set HOME env variable in profile.d' do
        expect(profile_d_script).to include('export HOME=/app')
      end

      it 'does not set NugetPackageRoot env variable in profile.d' do
        expect(profile_d_script).to_not include('export NugetPackageRoot=/app/.nuget/packages/')
      end

      it 'set LD_LIBRARY_PATH in profile.d' do
        expect(profile_d_script).to include('export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/ld_library_path:$HOME/libunwind/lib')
      end

      it 'add Dotnet CLI to the PATH in profile.d' do
        expect(profile_d_script).to include('$HOME/.dotnet')
      end

      it 'start command does not contain any exports' do
        expect(web_process).not_to include('export')
      end

      it "runs 'dotnet run' for project" do
        expect(web_process).to match('dotnet run --project foo')
      end

      context 'LD_LIBRARY_PATH specifies a custom library path' do
        before do
          ENV['LD_LIBRARY_PATH'] = '$HOME/my_custom_library/'
        end

        it 'appends the custom library path to LD_LIBRARY_PATH in profile.d' do
          expect(profile_d_script).to include('export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/ld_library_path:$HOME/libunwind/lib:$HOME/my_custom_library/')
        end
      end

      context 'multiple directories contain project.json files but no .deployment file exists' do
        let(:proj1) { File.join(build_dir, 'src', 'foo').tap { |f| FileUtils.mkdir_p(f) } }
        let(:proj2) { File.join(build_dir, 'src', 'proj2').tap { |f| FileUtils.mkdir_p(f) } }
        let(:proj3) { File.join(build_dir, 'src', 'proj3').tap { |f| FileUtils.mkdir_p(f) } }

        before do
          File.open(File.join(proj1, 'project.json'), 'w') do |f|
            f.write '{"dependencies": {"dep1": "whatever"}}'
          end
          File.open(File.join(proj2, 'project.json'), 'w') do |f|
            f.write '{"dependencies": {"dep1": "whatever"}}'
          end
          File.open(File.join(proj3, 'project.json'), 'w') do |f|
            f.write '{"dependencies": {"dep1": "whatever"}}'
          end
        end

        context '.deployment file exists' do
          before do
            File.open(File.join(build_dir, '.deployment'), 'w') do |f|
              f.write "[config]\n"
              f.write 'project=src/proj2'
            end
          end
          let(:web_process) do
            yml = YAML.load(subject.release(build_dir))
            yml.fetch('default_process_types').fetch('web')
          end
          it 'runs the project specified in the .deployment file' do
            expect(web_process).to match('dotnet run --project src/proj2')
          end
        end
      end
    end

    context '*.csproj exists' do
      let(:is_sdk_project_json) { false }
      let(:is_sdk_msbuild)      { true }

      let(:proj1) { File.join(build_dir, 'foo').tap { |f| Dir.mkdir(f) } }
      let(:csproj) { '<xml>anything</xml>' }

      let(:profile_d_script) do
        allow_any_instance_of(AspNetCoreBuildpack::DotnetSdkInstaller).to receive(:cached?).and_return(true)
        allow_any_instance_of(AspNetCoreBuildpack::LibunwindInstaller).to receive(:cached?).and_return(true)
        subject.release(build_dir)
        IO.read(File.join(build_dir, '.profile.d', 'startup.sh'))
      end

      let(:web_process) do
        yml = YAML.load(subject.release(build_dir))
        yml.fetch('default_process_types').fetch('web')
      end

      before do
        File.open(File.join(proj1, 'foo.csproj'), 'w') do |f|
          f.write csproj
        end
      end

      it 'set HOME env variable in profile.d' do
        expect(profile_d_script).to include('export HOME=/app')
      end

      it 'sets NugetPackageRoot env variable in profile.d' do
        expect(profile_d_script).to include('export NugetPackageRoot=/app/.nuget/packages/')
      end

      it 'set LD_LIBRARY_PATH in profile.d' do
        expect(profile_d_script).to include('export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/ld_library_path:$HOME/libunwind/lib')
      end

      it 'add Dotnet CLI to the PATH in profile.d' do
        expect(profile_d_script).to include('$HOME/.dotnet')
      end

      it 'start command does not contain any exports' do
        expect(web_process).not_to include('export')
      end

      it "runs 'dotnet run' for project" do
        expect(web_process).to match('dotnet run --project foo/foo.csproj')
      end

      context 'LD_LIBRARY_PATH specifies a custom library path' do
        before do
          ENV['LD_LIBRARY_PATH'] = '$HOME/my_custom_library/'
        end

        it 'appends the custom library path to LD_LIBRARY_PATH in profile.d' do
          expect(profile_d_script).to include('export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/ld_library_path:$HOME/libunwind/lib:$HOME/my_custom_library/')
        end
      end

      context 'multiple directories contain project.json files but no .deployment file exists' do
        let(:proj1) { File.join(build_dir, 'src', 'foo').tap { |f| FileUtils.mkdir_p(f) } }
        let(:proj2) { File.join(build_dir, 'src', 'proj2').tap { |f| FileUtils.mkdir_p(f) } }
        let(:proj3) { File.join(build_dir, 'src', 'proj3').tap { |f| FileUtils.mkdir_p(f) } }

        before do
          File.open(File.join(proj1, 'foo.csproj'), 'w') do |f|
            f.write '<xml>anything</xml>'
          end
          File.open(File.join(proj2, 'proj2.csproj'), 'w') do |f|
            f.write '<xml>anything</xml>'
          end
          File.open(File.join(proj3, 'proj3.csproj'), 'w') do |f|
            f.write '<xml>anything</xml>'
          end
        end

        context '.deployment file exists' do
          before do
            File.open(File.join(build_dir, '.deployment'), 'w') do |f|
              f.write "[config]\n"
              f.write 'project=src/proj2/proj2.csproj'
            end
          end
          let(:web_process) do
            yml = YAML.load(subject.release(build_dir))
            yml.fetch('default_process_types').fetch('web')
          end
          it 'runs the project specified in the .deployment file' do
            expect(web_process).to match('dotnet run --project src/proj2')
          end
        end
      end
    end
  end
end
