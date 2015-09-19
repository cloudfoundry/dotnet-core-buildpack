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
require 'yaml'
require 'tmpdir'
require 'fileutils'
require_relative '../../../lib/buildpack.rb'

describe AspNet5Buildpack::ReleaseYmlWriter do
  let(:build_dir) do
    Dir.mktmpdir
  end

  let(:out) do
    double(:out)
  end

  describe 'the .profile.d script' do
    let(:web_dir) do
      File.join(build_dir, 'foo').tap { |f| Dir.mkdir(f) }
    end

    let(:project_json) do
      '{"commands": {"kestrel": "whatever"}}'
    end

    before do
      File.open(File.join(web_dir, 'project.json'), 'w') do |f|
        f.write project_json
      end
    end

    let(:profile_d_script) do
      subject.write_release_yml(build_dir, out)
      IO.read(File.join(build_dir, '.profile.d', 'startup.sh'))
    end

    it 'should add /app/mono/bin to the PATH' do
      expect(profile_d_script).to include('export PATH=$HOME/mono/bin:$PATH;')
    end

    it 'should set HOME to /app (so that dependencies are picked up from /app/.dnx)' do
      expect(profile_d_script).to include('export HOME=/app')
    end

    it 'make sure dlopen can access libuv.so.1' do
      expect(profile_d_script).to include('export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/libuv/lib')
    end

    it 'should source dnvm script' do
      expect(profile_d_script).to include('source $HOME/.dnx/dnvm/dnvm.sh')
    end

    it 'should add the runtime to the PATH' do
      expect(profile_d_script).to include('dnvm use default')
    end
  end

  describe 'the release yml' do
    let(:web_process) do
      subject.write_release_yml(build_dir, out)
      yml = YAML.load_file(File.join(build_dir, 'aspnet5-buildpack-release.yml'))
      yml.fetch('default_process_types').fetch('web')
    end

    context 'when there are no directories containing a project.json' do
      it 'should raise an error because dnu/dnx will not work' do
        expect { subject.write_release_yml(build_dir, out) }.to raise_error(/No application found/)
      end
    end

    context 'when there is a directory with a project.json file' do
      let(:web_dir) do
        File.join(build_dir, 'foo').tap { |f| Dir.mkdir(f) }
      end

      let(:project_json) do
        '{"commands": {"kestrel": "whatever"}}'
      end

      before do
        File.open(File.join(web_dir, 'project.json'), 'w') do |f|
          f.write project_json
        end
      end

      it 'does not contain any exports (these should be done via .profile.d script)' do
        expect(web_process).not_to include('export')
      end

      context 'and the project.json does not contain a kestrel command' do
        let(:project_json) do
          '{"commands": {"web": "whatever"}}'
        end

        it 'should raise an error because dnx will not work' do
          expect { subject.write_release_yml(build_dir, out) }.to raise_error(/No kestrel command found in foo/)
        end
      end

      context 'and the project.json contains a kestrel command' do
        let(:project_json) do
          '{"commands": {"kestrel": "whatever"}}'
        end

        it "runs 'dnx kestrel'" do
          expect(web_process).to match('dnx --project foo kestrel')
        end
      end
    end

    context 'when there are multiple directories containing project.json files' do
      let(:proj1) do
        File.join(build_dir, 'src', 'proj1').tap { |f| FileUtils.mkdir_p(f) }
      end

      let(:proj2) do
        File.join(build_dir, 'src', 'proj2').tap { |f| FileUtils.mkdir_p(f) }
      end

      before do
        File.open(File.join(proj1, 'project.json'), 'w') do |f|
          f.write '{"commands": {"migrate": "whatever"}}'
        end
        File.open(File.join(proj2, 'project.json'), 'w') do |f|
          f.write '{"commands": {"kestrel": "whatever"}}'
        end
      end

      it "runs 'dnx kestrel'" do
        expect(web_process).to match('dnx --project src/proj2 kestrel')
      end
    end

    context 'when project.json is in the base app directory' do
      before do
        File.open(File.join(build_dir, 'project.json'), 'w') do |f|
          f.write '{"commands": {"kestrel": "whatever"}}'
        end
      end

      it "runs 'dnx kestrel'" do
        expect(web_process).to match('dnx --project . kestrel')
      end
    end

    context 'when there is a packages directory' do
      before do
        FileUtils.mkdir_p(File.join(build_dir, 'approot', 'packages'))
        File.open(File.join(build_dir, 'kestrel'), 'w') { |f| f.write 'x' }
        File.open(File.join(build_dir, 'approot', 'project.json'), 'w') do |f|
          f.write '{"commands": {"kestrel": "whatever"}}'
        end
      end

      it 'runs the kestrel script' do
        expect(web_process).to match('./kestrel')
      end
    end
  end
end
