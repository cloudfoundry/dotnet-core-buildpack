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

  describe 'the release yml' do
    let(:yml_path) do
      File.join(build_dir, 'aspnet5-buildpack-release.yml')
    end

    let(:yml) do
      subject.write_release_yml(build_dir, out)
      YAML.load_file(yml_path)
    end

    let(:profile_d_script) do
      subject.write_release_yml(build_dir, out)
      IO.read(File.join(build_dir, '.profile.d', 'startup.sh'))
    end

    describe 'the .profile.d script' do
      let(:web_dir) do
        File.join(build_dir, 'foo').tap { |f| Dir.mkdir(f) }
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

      it 'should re-run package restore' do
        expect(profile_d_script).to include('dnu restore')
      end
    end

    describe 'the web process type' do
      let(:web_process) do
        yml.fetch('default_process_types').fetch('web')
      end

      context 'when there are no directories containing a project.json' do
        it 'should work (the user might be using a custom start command)' do
          expect(out).not_to receive(:fail)
          subject.write_release_yml(build_dir, out)
          expect(File).to exist(yml_path)
        end
      end

      context 'when there is a directory with a project.json file containing a BOM' do
        let(:web_dir) do
          File.join(build_dir, 'foo').tap { |f| Dir.mkdir(f) }
        end

        let(:project_json) do
          '{}'
        end

        before do
          File.open(File.join(web_dir, 'project.json'), 'w') do |f|
            f.write "\uFEFF"
            f.write project_json
          end
        end

        it 'writes a release yml' do
          subject.write_release_yml(build_dir, out)
          expect(File).to exist(File.join(build_dir, 'aspnet5-buildpack-release.yml'))
        end
      end

      context 'when there is a directory with a project.json file' do
        let(:web_dir) do
          File.join(build_dir, 'foo').tap { |f| Dir.mkdir(f) }
        end

        let(:project_json) do
          '{}'
        end

        before do
          File.open(File.join(web_dir, 'project.json'), 'w') do |f|
            f.write project_json
          end
        end

        it 'writes a release yml' do
          subject.write_release_yml(build_dir, out)
          expect(File).to exist(File.join(build_dir, 'aspnet5-buildpack-release.yml'))
        end

        it 'contains a web process type' do
          expect(yml).to have_key('default_process_types')
          expect(yml.fetch('default_process_types')).to have_key('web')
        end

        it 'does not contain any exports (these should be done via .profile.d script)' do
          expect(yml).to have_key('default_process_types')
          expect(yml['default_process_types']['web']).not_to include('export')
        end

        context 'and the project.json contains a cf-web command' do
          let(:project_json) do
            '{"commands": {"cf-web": "whatever"}}'
          end

          it 'changes directory to that directory' do
            expect(web_process).to match('cd foo;')
          end

          it "runs 'dnx . cf-web'" do
            expect(web_process).to match('dnx . cf-web')
          end

          context 'and if the cf-web command is empty' do
            let(:project_json) do
              '{"commands": {"cf-web": ""}}'
            end

            it 'sets it to serve Kestrel' do
              subject.write_release_yml(build_dir, out)

              json = JSON.parse(IO.read(File.join(web_dir, 'project.json')))
              expect(json['commands']['cf-web']).to match('Microsoft.AspNet.Hosting --server Kestrel')
            end
          end
        end

        context 'and the project.json does not contain a cf-web command' do
          it 'adds cf-web command to project.json' do
            subject.write_release_yml(build_dir, out)

            json = JSON.parse(IO.read(File.join(web_dir, 'project.json')))
            expect(json).to have_key('commands')
            expect(json['commands']).to have_key('cf-web')
            expect(json['commands']['cf-web']).to match('Microsoft.AspNet.Hosting --server Kestrel')
          end

          context 'when Kestrel dependency exists' do
            let(:project_json) do
              '{ "dependencies" : { "Kestrel" : "345" } }'
            end

            it 'leaves it alone' do
              subject.write_release_yml(build_dir, out)

              json = JSON.parse(IO.read(File.join(web_dir, 'project.json')))
              expect(json['dependencies']['Kestrel']).to match('345')
            end
          end

          context 'when Kestrel dependency does not exist' do
            before do
              json = '{ "sdk": { "version": "1.0.0-beta1" } }'
              IO.write(File.join(build_dir, 'global.json'), json)
            end
            it 'adds Kestrel dependency to project.json' do
              subject.write_release_yml(build_dir, out)

              json = JSON.parse(IO.read(File.join(web_dir, 'project.json')))
              expect(json).to have_key('dependencies')
              expect(json['dependencies']).to have_key('Kestrel')
              expect(json['dependencies']['Kestrel']).to match('1.0.0-beta1')
            end
          end
        end
      end

      context 'when there are multiple directories with a project.json file' do
        let(:web_dir) do
          File.join(build_dir, 'foo-cfweb').tap { |f| Dir.mkdir(f) }
        end

        let(:other_dir) do
          File.join(build_dir, 'bar').tap { |f| Dir.mkdir(f) }
        end

        context 'and one contains a cf-web command' do
          before do
            File.open(File.join(other_dir, 'project.json'), 'w') do |f|
              f.write '{ "commands": { "web": "whatever" } }'
            end

            File.open(File.join(web_dir, 'project.json'), 'w') do |f|
              f.write '{ "commands": { "cf-web": "whatever" } }'
            end
          end

          it 'changes directory to that directory' do
            expect(web_process).to match('cd foo-cfweb;')
          end

          it "runs 'dnx . cf-web at correct host and port'" do
            expect(web_process).to match(%r{dnx . cf-web --server.urls http:\/\/\$\{VCAP_APP_HOST\}\:\$\{PORT\}})
          end
        end

        context 'and a .deployment file exists' do
          before do
            File.open(File.join(other_dir, 'project.json'), 'w') do |f|
              f.write '{ "commands": { "web": "whatever" } }'
            end

            File.open(File.join(web_dir, 'project.json'), 'w') do |f|
              f.write '{ "commands": { "web": "whatever" } }'
            end
          end

          it 'changes directory to that directory when a project is specified' do
            File.open(File.join(build_dir, '.deployment'), 'w') do |f|
              f.write("project = bar\n")
            end
            expect(web_process).to match('cd bar;')
          end

          it 'changes directory to a valid directory when an invalid project is specified' do
            File.open(File.join(build_dir, '.deployment'), 'w') do |f|
              f.write("project = dne\n")
            end
            expect(web_process).to match(/cd (foo-cfweb|bar);/)
          end

          it 'changes directory to a valid directory when no project is specified' do
            File.open(File.join(build_dir, '.deployment'), 'w') do |f|
              f.write("[config]\n")
            end
            expect(web_process).to match(/cd (foo-cfweb|bar);/)
          end
        end
      end
    end
  end
end
