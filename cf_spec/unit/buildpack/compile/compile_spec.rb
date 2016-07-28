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

describe AspNetCoreBuildpack::Compiler do
  subject(:compiler) do
    AspNetCoreBuildpack::Compiler.new(build_dir, cache_dir, libunwind_binary, dotnet_installer, dotnet, copier, out)
  end

  before do
    allow($stdout).to receive(:write)
  end

  let(:libunwind_binary) { double(:libunwind_binary, extract: nil) }
  let(:copier) { double(:copier, cp: nil) }
  let(:dotnet_installer) { double(:dotnet_installer, install: nil, should_install: true) }
  let(:dotnet) { double(:dotnet, restore: nil) }
  let(:build_dir) { Dir.mktmpdir }
  let(:cache_dir) { Dir.mktmpdir }

  let(:out) do
    double(:out, step: double(:unknown_step, succeed: nil)).tap do |out|
      allow(out).to receive(:warn)
    end
  end

  shared_examples 'step' do |expected_message, step|
    let(:step_out) do
      double(:step_out, succeed: nil).tap do |step_out|
        allow(out).to receive(:step).with(expected_message).and_return step_out
      end
    end

    it 'outputs step name' do
      expect(out).to receive(:step).with(expected_message)
      compiler.compile
    end

    it 'runs step' do
      expect(step_out).to receive(:succeed)
      compiler.compile
    end

    context 'step fails' do
      it 'prints helpful error' do
        allow(subject).to receive(step).and_raise 'fishfinger in the warp core'
        allow(out).to receive(:fail)
        allow(step_out).to receive(:fail)
        allow(out).to receive(:warn)
        expect(step_out).to receive(:fail).with(match(/fishfinger in the warp core/))
        expect(out).to receive(:fail).with(match(/#{expected_message} failed, fishfinger in the warp core/))
        expect { compiler.compile }.not_to raise_error
      end
    end
  end

  describe 'Steps' do
    describe 'Restoring Cache' do
      it_behaves_like 'step', 'Restoring files from buildpack cache', :restore_cache

      context 'cache does not exist' do
        it 'skips restore' do
          expect(copier).not_to receive(:cp).with(match(cache_dir), anything, anything)
          compiler.compile
        end

        it 'binary files extracted' do
          expect(libunwind_binary).to receive(:extract)
          compiler.compile
        end
      end

      context 'cache exists' do
        before(:each) do
          Dir.mkdir(File.join(cache_dir, '.nuget'))
          Dir.mkdir(File.join(cache_dir, 'libunwind'))
          Dir.mkdir(File.join(build_dir, 'libunwind'))
        end

        it 'copies files from cache to build dir' do
          expect(copier).to receive(:cp).with(File.join(cache_dir, '.nuget'), build_dir, anything)
          expect(copier).to receive(:cp).with(File.join(cache_dir, 'libunwind'), build_dir, anything)
          compiler.compile
        end

        it 'binary files not extracted' do
          expect(libunwind_binary).not_to receive(:extract)
          compiler.compile
        end
      end
    end

    describe 'Installing Dotnet CLI' do
      it_behaves_like 'step', 'Installing Dotnet CLI', :install_dotnet

      it 'installs dotnet cli' do
        expect(dotnet_installer).to receive(:install).with(build_dir, anything)
        compiler.compile
      end

      context 'when the app was published' do
        it 'skips installing Dotnet CLI' do
          allow(dotnet_installer).to receive(:should_install).and_return(false)
          expect(dotnet_installer).not_to receive(:install).with(build_dir, anything)
          compiler.compile
        end
      end
    end

    describe 'Restoring dependencies with Dotnet CLI' do
      it_behaves_like 'step', 'Restoring dependencies with Dotnet CLI', :restore_dependencies

      it 'runs dotnet restore' do
        expect(dotnet).to receive(:restore).with(build_dir, anything)
        compiler.compile
      end

      context 'when the app was published' do
        it 'skips running dotnet restore' do
          allow(dotnet_installer).to receive(:should_install).and_return(false)
          expect(dotnet).not_to receive(:restore)
          compiler.compile
        end
      end
    end

    describe 'Saving to buildpack cache' do
      it_behaves_like 'step', 'Saving to buildpack cache', :save_cache

      it 'copies files to cache dir' do
        expect(copier).to receive(:cp).with("#{build_dir}/libunwind", cache_dir, anything)
        compiler.compile
      end

      context 'when the cache already exists' do
        before(:each) do
          Dir.mkdir(File.join(build_dir, '.nuget'))
          Dir.mkdir(File.join(cache_dir, 'libunwind'))
        end

        it 'copies only .nuget to cache dir' do
          expect(copier).to receive(:cp).with("#{build_dir}/.nuget", cache_dir, anything)
          compiler.compile
        end
      end
    end
  end
end
