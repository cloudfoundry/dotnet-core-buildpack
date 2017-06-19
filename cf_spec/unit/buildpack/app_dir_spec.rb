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

describe AspNetCoreBuildpack::AppDir do
  let(:dir)                   { Dir.mktmpdir }
  let(:deps_dir)                   { Dir.mktmpdir }
  let(:deps_idx)                   { '55' }
  let(:app_uses_msbuild)      { false }
  let(:app_uses_project_json) { true }

  subject(:appdir) { described_class.new(dir, deps_dir, deps_idx) }

  before do
    allow(appdir).to receive(:msbuild?).and_return(app_uses_msbuild)
    allow(appdir).to receive(:project_json?).and_return(app_uses_project_json)
  end

  context 'with multiple .csproj projects' do
    let(:app_uses_msbuild)      { true }
    let(:app_uses_project_json) { false }

    let(:proj1) { File.join(dir, 'src', 'proj1').tap { |f| FileUtils.mkdir_p(f) } }
    let(:proj2) { File.join(dir, 'src', 'föö').tap { |f| FileUtils.mkdir_p(f) } }
    let(:nuget) { File.join(dir, '.hidden', 'dep').tap { |f| FileUtils.mkdir_p(f) } }

    before do
      File.open(File.join(proj1, 'proj1.fsproj'), 'w') do |f|
        f.write 'a fsproj file'
      end
      File.open(File.join(proj2, 'föö.csproj'), 'w') do |f|
        f.write 'an csproj file'
      end
      File.open(File.join(nuget, 'dep.csproj'), 'w') do |f|
        f.write 'a third csproj file'
      end
    end

    it 'finds all MSBuild files from non-hidden directories' do
      expect(subject.msbuild_projects).to match_array([Pathname.new('src/proj1/proj1.fsproj'), Pathname.new('src/föö/föö.csproj')])
    end

    context '.deployment file exists' do
      context 'and specifies an existing project without a ./ at the start of the path' do
        before do
          File.open(File.join(dir, '.deployment'), 'w') do |f|
            f.write("[config]\n")
            f.write("project = src/föö/föö.csproj\n")
          end
        end

        it 'finds specified project' do
          expect(subject.deployment_file_project).to eq(Pathname.new('src/föö/föö.csproj'))
        end
      end
    end
    context '.deployment file exists' do
      context 'and specifies an existing project with a ./ at the start of the path' do
        before do
          File.open(File.join(dir, '.deployment'), 'w') do |f|
            f.write("[config]\n")
            f.write("project = ./src/föö/föö.csproj\n")
          end
        end

        it 'finds specified project' do
          expect(subject.deployment_file_project).to eq(Pathname.new('src/föö/föö.csproj'))
        end
      end
    end
  end

  context 'with multiple projects with project.json' do
    let(:proj1) { File.join(dir, 'src', 'proj1').tap { |f| FileUtils.mkdir_p(f) } }
    let(:proj2) { File.join(dir, 'src', 'föö').tap { |f| FileUtils.mkdir_p(f) } }
    let(:nuget) { File.join(dir, '.hidden', 'dep').tap { |f| FileUtils.mkdir_p(f) } }

    before do
      File.open(File.join(proj1, 'project.json'), 'w') do |f|
        f.write '{ "commands": { "web1": "whatever", "web2": "whatever" } }'
      end
      File.open(File.join(proj2, 'project.json'), 'w') do |f|
        f.write "\uFEFF"
        f.write '{ "commands": { "web": "whatever" } }'
      end
      File.open(File.join(nuget, 'project.json'), 'w') do |f|
        f.write '{ "commands": { "web": "whatever" } }'
      end
    end

    it 'finds all project.json files from non-hidden directories' do
      expect(subject.with_project_json).to match_array([Pathname.new('src/proj1'), Pathname.new('src/föö')])
    end

    context '.deployment file exists' do
      context 'and specifies an existing project without a ./ at the start of path' do
        before do
          File.open(File.join(dir, '.deployment'), 'w') do |f|
            f.write("[config]\n")
            f.write("project = src/föö\n")
          end
        end

        it 'finds specified project' do
          expect(subject.deployment_file_project).to eq(Pathname.new('src/föö'))
        end
      end

      context 'and specifies an existing project with a ./ at the start of path' do
        before do
          File.open(File.join(dir, '.deployment'), 'w') do |f|
            f.write("[config]\n")
            f.write("project = ./src/föö\n")
          end
        end

        it 'finds specified project' do
          expect(subject.deployment_file_project).to eq(Pathname.new('src/föö'))
        end
      end

      context 'and specifies an existing .xproj file' do
        before do
          File.open(File.join(dir, '.deployment'), 'w') do |f|
            f.write("[config]\n")
            f.write("project = src/föö/project.xproj\n")
          end
        end

        it 'finds specified project' do
          expect(subject.deployment_file_project).to eq(Pathname.new('src/föö'))
        end
      end

      context 'and specifies an existing .csproj file' do
        before do
          File.open(File.join(dir, '.deployment'), 'w') do |f|
            f.write("[config]\n")
            f.write("project = src/föö/project.csproj\n")
          end
        end

        it 'finds specified project' do
          expect(subject.deployment_file_project).to eq(Pathname.new('src/föö'))
        end
      end

      context 'and does not specify a project' do
        before do
          File.open(File.join(dir, '.deployment'), 'w') do |f|
            f.write("[config]\n")
            f.write("some_other_key = src/föö\n")
          end
        end

        it 'throws a helpful error about needing a project key' do
          expect{ subject.deployment_file_project }.to raise_error(AspNetCoreBuildpack::DeploymentConfigError,
                                                                   /Invalid .deployment file: must have project key/)
        end
      end

      context 'and does not specifies more than one project' do
        before do
          File.open(File.join(dir, '.deployment'), 'w') do |f|
            f.write("[config]\n")
            f.write("project = src/föö\n")
            f.write("project = src/fööbar\n")
          end
        end

        it 'throws a helpful error about only allowing one project key' do
          expect{ subject.deployment_file_project }.to raise_error(AspNetCoreBuildpack::DeploymentConfigError,
                                                                   /Invalid .deployment file: must only contain one project key/)
        end
      end

      context 'and specifies a non-existent project' do
        before do
          File.open(File.join(dir, '.deployment'), 'w') do |f|
            f.write("[config]\n")
            f.write("project = dne\n")
          end
        end

        it 'does not find a project' do
          expect(subject.deployment_file_project).to be_nil
        end
      end

      context 'contains a byte order mark' do
        before do
          File.open(File.join(dir, '.deployment'), 'w') do |f|
            f.write("\uFEFF")
            f.write("[config]\n")
          end
        end

        context 'but does not specify a project' do
          it 'throws a helpful error about needing a project key' do
            expect{ subject.deployment_file_project }.to raise_error(AspNetCoreBuildpack::DeploymentConfigError, /Invalid .deployment file: must have project key/)
          end
        end

        context 'and specifies an existing project' do
          before do
            File.open(File.join(dir, '.deployment'), 'a') do |f|
              f.write('project = src/proj1')
            end
          end

          it 'finds specified project' do
            expect(subject.deployment_file_project).to eq(Pathname.new('src/proj1'))
          end
        end
      end
    end

    context 'no .deployment file exists' do
      it 'does not find a project' do
        expect(subject.deployment_file_project).to be_nil
      end

      it 'raises an error to tell the user that they need a .deployment file' do
        error_message = 'Multiple paths contain a project.json file, but no .deployment file was used'
        expect { subject.main_project_path }.to raise_error(error_message)
      end
    end

    context '*.runtimeconfig.json file exists in published app' do
      before do
        File.open(File.join(dir, 'proj1.runtimeconfig.json'), 'w') { |f| f.write 'x' }
      end

      it 'determines app name based on runtimeconfig.json file name' do
        expect(subject.published_project).to match('proj1')
      end
    end
  end
end
