# Encoding: utf-8
# ASP.NET 5 Buildpack
# Copyright 2014-2015 the original author or authors.
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

require 'rspec'
require 'tmpdir'
require 'fileutils'
require_relative '../../lib/buildpack.rb'

describe AspNet5Buildpack::AppDir do
  let(:dir) { Dir.mktmpdir }
  subject(:appdir) { AspNet5Buildpack::AppDir.new(dir) }

  context 'with multiple projects' do
    let(:proj1) { File.join(dir, 'src', 'proj1').tap { |f| FileUtils.mkdir_p(f) } }
    let(:proj2) { File.join(dir, 'src', 'proj2').tap { |f| FileUtils.mkdir_p(f) } }
    let(:dnx) { File.join(dir, '.dnx', 'dep').tap { |f| FileUtils.mkdir_p(f) } }

    before do
      File.open(File.join(proj1, 'project.json'), 'w') do |f|
        f.write '{ "commands": { "web1": "whatever", "web2": "whatever" } }'
      end
      File.open(File.join(proj2, 'project.json'), 'w') do |f|
        f.write "\uFEFF"
        f.write '{ "commands": { "web": "whatever" } }'
      end
      File.open(File.join(dnx, 'project.json'), 'w') do |f|
        f.write '{ "commands": { "web": "whatever" } }'
      end
    end

    it 'finds all project.json files from non-hidden directories' do
      expect(appdir.with_project_json).to match_array([Pathname.new('src/proj1'), Pathname.new('src/proj2')])
    end

    it 'finds project paths where project.json files have specific commands' do
      expect(appdir.with_command('web')).to match_array([Pathname.new('src/proj2')])
    end

    it 'does not find project paths where no project.json files have specific command' do
      expect(appdir.with_command('fakecmd')).to match_array([])
    end

    it 'reads commands from project.json files' do
      expect(appdir.commands('src/proj1')).to eq('web1' => 'whatever', 'web2' => 'whatever')
    end

    it 'reads commands from project.json files with byte-order marks' do
      expect(appdir.commands('src/proj2')).to eq('web' => 'whatever')
    end

    context '.deployment file specifies an existing project' do
      before do
        File.open(File.join(dir, '.deployment'), 'w') do |f|
          f.write("project = src/proj1\n")
        end
      end

      it 'finds specified project' do
        expect(appdir.deployment_file_project).to eq(Pathname.new('src/proj1'))
      end
    end

    context 'no .deployment file exists' do
      it 'does not find a project' do
        expect(appdir.deployment_file_project).to be_nil
      end
    end

    context '.deployment file specifies a non-existent project' do
      before do
        File.open(File.join(dir, '.deployment'), 'w') do |f|
          f.write("project = dne\n")
        end
      end

      it 'does not find a project' do
        expect(appdir.deployment_file_project).to be_nil
      end
    end

    context '.deployment file exists but does not specify a project' do
      before do
        File.open(File.join(dir, '.deployment'), 'w') do |f|
          f.write("[config]\n")
        end
      end

      it 'does not find a project' do
        expect(appdir.deployment_file_project).to be_nil
      end
    end
  end
end
