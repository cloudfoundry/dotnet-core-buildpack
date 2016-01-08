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

describe AspNet5Buildpack::Releaser do
  let(:build_dir) { Dir.mktmpdir }

  describe '#release' do
    context 'project.json does not exist' do
      it 'raises an error because dnu/dnx commands will not work' do
        expect { subject.release(build_dir) }.to raise_error(/No application found/)
      end
    end

    context 'project.json exists' do
      let(:proj1) { File.join(build_dir, 'foo').tap { |f| Dir.mkdir(f) } }
      let(:project_json) { '{"commands": {"kestrel": "whatever"}}' }

      let(:profile_d_script) do
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

      it 'set LD_LIBRARY_PATH in profile.d' do
        expect(profile_d_script).to include('export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/libuv/lib')
      end

      it 'source dnvm.sh script in profile.d' do
        expect(profile_d_script).to include('source $HOME/.dnx/dnvm/dnvm.sh')
      end

      it 'add DNX to the PATH in profile.d' do
        expect(profile_d_script).to include('dnvm use default')
      end

      it 'start command does not contain any exports' do
        expect(web_process).not_to include('export')
      end

      it "runs 'dnx kestrel' for project" do
        expect(web_process).to match('dnx --project foo kestrel')
      end

      context 'project.json does not contain a kestrel or web command' do
        let(:project_json) { '{"commands": {"notkestrelorweb": "whatever"}}' }

        it 'raises an error because start command will not work' do
          expect { subject.release(build_dir) }.to raise_error(/No kestrel or web command found in foo/)
        end
      end

      context 'project.json contains a kestrel command' do
        it "runs 'dnx kestrel' for project" do
          expect(web_process).to match('dnx --project foo kestrel')
        end
      end

      context 'project.json contains a web command' do
        let(:project_json) { '{"commands": {"web": "whatever"}}' }

        it "runs 'dnx web' for project" do
          expect(web_process).to match('dnx --project foo web')
        end
      end

      context 'multiple directories contain project.json files' do
        let(:proj2) { File.join(build_dir, 'src', 'proj2').tap { |f| FileUtils.mkdir_p(f) } }

        before do
          File.open(File.join(proj1, 'project.json'), 'w') do |f|
            f.write '{"commands": {"migrate": "whatever"}}'
          end
          File.open(File.join(proj2, 'project.json'), 'w') do |f|
            f.write '{"commands": {"kestrel": "whatever"}}'
          end
        end

        it "runs 'dnx kestrel' for project with kestrel command" do
          expect(web_process).to match('dnx --project src/proj2 kestrel')
        end
      end

      context 'project.json is in a published app' do
        before do
          FileUtils.mkdir_p(File.join(build_dir, 'approot', 'packages'))
          File.open(File.join(build_dir, 'approot', 'kestrel'), 'w') { |f| f.write 'x' }
          File.open(File.join(build_dir, 'approot', 'project.json'), 'w') do |f|
            f.write '{"commands": {"kestrel": "whatever"}}'
          end
        end

        it 'runs kestrel script' do
          expect(web_process).to match('approot/kestrel')
        end
      end
    end
  end
end
