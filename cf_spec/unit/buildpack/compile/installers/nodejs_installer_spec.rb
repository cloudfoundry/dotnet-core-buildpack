# Encoding: utf-8
# ASP.NET Core Buildpack
# Copyright 2016 the original author or authors.
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

describe AspNetCoreBuildpack::NodeJsInstaller do
  let(:dir) { Dir.mktmpdir }
  let(:cache_dir) { Dir.mktmpdir }
  let(:shell) { double(:shell, env: {}) }
  let(:out) { double(:out) }
  let(:self_contained_app_dir) { double(:self_contained_app_dir, published_project: 'project1') }
  let(:app_dir) { double(:app_dir, published_project: false, with_project_json: %w(['project1', 'project2'])) }
  subject(:installer) { described_class.new(dir, cache_dir, shell) }

  describe '#cached?' do
    context 'cache directory exists in the buildpack cache' do
      before do
        FileUtils.mkdir_p(File.join(cache_dir, '.node', 'node-v6.7.0-linux-x64', 'bin'))
      end

      it 'returns true' do
        expect(subject.send(:cached?)).to be_truthy
      end
    end

    context 'cache directory does not exist in the build directory' do
      it 'returns false' do
        expect(subject.send(:cached?)).not_to be_truthy
      end
    end
  end

  describe '#install' do
    it 'downloads file with compile-extensions' do
      allow(shell).to receive(:exec).and_return(0)
      expect(shell).to receive(:exec) do |*args|
        cmd = args.first
        expect(cmd).to match(/download_dependency/)
        expect(cmd).to match(/tar/)
      end
      expect(out).to receive(:print).with(/Node.js version/)
      subject.install(out)
    end

    context 'another version of Node.js exists in cache' do
      before do
        FileUtils.mkdir_p(File.join(dir, subject.cache_dir, 'other_version_of_node'))
      end

      it 'clears any old versions from the cache folder' do
        expect(out).to receive(:print).with(anything)
        allow(shell).to receive(:exec).and_return(0)
        expect(shell).to receive(:exec) do |*args|
          cmd = args.first
          expect(cmd).to match(anything)
          expect(cmd).to match(anything)
        end
        subject.install(out)

        expect(File.exist?(File.join(dir, subject.cache_dir, 'other_version_of_node'))).not_to be_truthy
      end
    end
  end

  describe '#should_install' do
    context 'app is self-contained' do
      it 'returns false' do
        expect(subject.should_install(self_contained_app_dir)).not_to be_truthy
      end
    end

    context 'app is not self-contained' do
      before do
        FileUtils.mkdir_p(File.join(dir, 'src', 'project1'))
      end

      context 'scripts section exists' do
        context 'has both npm and bower commands' do
          before do
            FileUtils.mkdir_p(File.join(dir, 'src', 'project1'))
            File.open(File.join(dir, 'src', 'project1', 'project.json'), 'w') { |f| f.write('{"scripts": { "precompile":["npm install", "bower install"] }}') }
          end

          it 'returns true' do
            expect(subject.should_install(app_dir)).to be_truthy
          end
        end

        context 'has only npm command' do
          before do
            File.open(File.join(dir, 'src', 'project1', 'project.json'), 'w') { |f| f.write('{"scripts": { "precompile": "npm install" }}') }
          end

          it 'returns true' do
            expect(subject.should_install(app_dir)).to be_truthy
          end
        end

        context 'has only bower command' do
          before do
            File.open(File.join(dir, 'src', 'project1', 'project.json'), 'w') { |f| f.write('{"scripts": { "precompile": "bower install" }}') }
          end

          it 'returns true' do
            expect(subject.should_install(app_dir)).to be_truthy
          end
        end

        context 'has no bower or npm command' do
          before do
            File.open(File.join(dir, 'src', 'project1', 'project.json'), 'w') { |f| f.write('{"scripts": { "precompile": "minify js" }}') }
          end

          it 'returns false' do
            expect(subject.should_install(app_dir)).not_to be_truthy
          end
        end
      end

      it 'returns false when scripts section does not exist' do
        allow(subject).to receive(:get_scripts_section).with(anything).and_return(nil)
        expect(subject.should_install(app_dir)).not_to be_truthy
      end
    end
  end
end
