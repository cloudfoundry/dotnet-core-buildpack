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
require_relative '../../../lib/buildpack.rb'

describe AspNet5Buildpack::Compiler do
  subject(:compiler) do
    AspNet5Buildpack::Compiler.new(build_dir, cache_dir, libuv_binary, libunwind_binary, dnvm_installer, dnx_installer, dnu, copier, out)
  end

  before do
    allow($stdout).to receive(:write)
  end

  let(:libuv_binary) { double(:libuv_binary, extract: nil) }
  let(:libunwind_binary) { double(:libunwind_binary, extract: nil) }
  let(:copier) { double(:copier, cp: nil) }
  let(:dnvm_installer) { double(:dnvm_installer, install: nil) }
  let(:dnx_installer) { double(:dnx_installer, install: nil) }
  let(:dnu) { double(:dnu, restore: nil) }
  let(:build_dir) { Dir.mktmpdir }
  let(:cache_dir) { Dir.mktmpdir }

  let(:out) do
    double(:out, step: double(:unknown_step, succeed: nil)).tap do |out|
      allow(out).to receive(:warn)
    end
  end

  it 'prints experimental warning message' do
    expect(out).to receive(:warn).with(match('experimental'))
    compiler.compile
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
        expect(out).to receive(:warn).with(match('experimental'))
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
          expect(libuv_binary).to receive(:extract)
          expect(libunwind_binary).to receive(:extract)
          compiler.compile
        end
      end

      context 'cache exists' do
        before(:each) do
          Dir.mkdir(File.join(cache_dir, '.dnx'))
          Dir.mkdir(File.join(cache_dir, 'libuv'))
          Dir.mkdir(File.join(cache_dir, 'libunwind'))
          Dir.mkdir(File.join(build_dir, 'libuv'))
          Dir.mkdir(File.join(build_dir, 'libunwind'))
        end

        it 'copies files from cache to build dir' do
          expect(copier).to receive(:cp).with(File.join(cache_dir, '.dnx'), build_dir, anything)
          expect(copier).to receive(:cp).with(File.join(cache_dir, 'libuv'), build_dir, anything)
          expect(copier).to receive(:cp).with(File.join(cache_dir, 'libunwind'), build_dir, anything)
          compiler.compile
        end

        it 'binary files not extracted' do
          expect(libuv_binary).not_to receive(:extract)
          expect(libunwind_binary).not_to receive(:extract)
          compiler.compile
        end
      end
    end

    describe 'Installing DNVM' do
      it_behaves_like 'step', 'Installing DNVM', :install_dnvm

      it 'installs dnvm' do
        expect(dnvm_installer).to receive(:install).with(build_dir, anything)
        compiler.compile
      end

      context 'when the app was published with DNX' do
        before do
          FileUtils.mkdir_p(File.join(build_dir, 'approot', 'runtimes'))
        end

        it 'skips installing DNVM' do
          expect(dnvm_installer).not_to receive(:install)
          compiler.compile
        end
      end
    end

    describe 'Installing DNX with DNVM' do
      it_behaves_like 'step', 'Installing DNX with DNVM', :install_dnx

      it 'installs dnx' do
        expect(dnx_installer).to receive(:install).with(build_dir, anything)
        compiler.compile
      end

      context 'when the app was published with DNX' do
        before do
          FileUtils.mkdir_p(File.join(build_dir, 'approot', 'runtimes'))
        end

        it 'skips installing DNX' do
          expect(dnx_installer).not_to receive(:install)
          compiler.compile
        end
      end
    end

    describe 'Restoring dependencies with DNU' do
      it_behaves_like 'step', 'Restoring dependencies with DNU', :restore_dependencies

      it 'runs dnu restore' do
        expect(dnu).to receive(:restore).with(build_dir, anything)
        compiler.compile
      end

      context 'when the app was published with NuGet packages' do
        before do
          FileUtils.mkdir_p(File.join(build_dir, 'approot', 'packages'))
        end

        it 'skips running dnu restore' do
          expect(dnu).not_to receive(:restore)
          compiler.compile
        end
      end
    end

    describe 'Saving to buildpack cache' do
      it_behaves_like 'step', 'Saving to buildpack cache', :save_cache

      it 'copies files to cache dir' do
        expect(copier).to receive(:cp).with("#{build_dir}/libuv", cache_dir, anything)
        expect(copier).to receive(:cp).with("#{build_dir}/libunwind", cache_dir, anything)
        compiler.compile
      end

      context 'when the cache already exists' do
        before(:each) do
          Dir.mkdir(File.join(build_dir, '.dnx'))
          Dir.mkdir(File.join(cache_dir, '.dnx'))
          Dir.mkdir(File.join(cache_dir, 'libuv'))
          Dir.mkdir(File.join(cache_dir, 'libunwind'))
        end

        it 'copies only .dnx to cache dir' do
          expect(copier).to receive(:cp).with("#{build_dir}/.dnx", cache_dir, anything)
          compiler.compile
        end
      end
    end
  end
end
