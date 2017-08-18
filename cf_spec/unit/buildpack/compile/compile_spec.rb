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
    described_class.new(build_dir, cache_dir, deps_dir, deps_idx, copier, installer.descendants, out)
  end

  before do
    allow($stdout).to receive(:write)
  end

  let(:installer) { double(:installer, descendants: [libunwind_installer]) }
  let(:libunwind_installer) do
    double(:libunwind_installer, install: nil).tap do |libunwind_installer|
      allow(libunwind_installer).to receive(:install_description)
      allow(libunwind_installer).to receive(:cache_dir).and_return('libunwind')
      allow(libunwind_installer).to receive(:should_install).and_return(true)
      allow(libunwind_installer).to receive(:name).and_return('libunwind')
      allow(libunwind_installer).to receive(:create_links)
    end
  end

  let(:copier) { double(:copier, cp: nil) }
  let(:build_dir) { Dir.mktmpdir }
  let(:cache_dir) { Dir.mktmpdir }
  let(:deps_dir) { Dir.mktmpdir }
  let(:deps_idx) { '0' }

  let(:out) do
    double(:out, step: double(:unknown_step, succeed: nil, print: nil)).tap do |out|
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
      allow(libunwind_installer).to receive(:cached?)
      subject.supply
    end

    it 'runs step' do
      allow(step_out).to receive(:print)
      expect(step_out).to receive(:succeed)
      allow(libunwind_installer).to receive(:cached?)
      subject.supply
    end

    context 'step fails' do
      it 'prints helpful error' do
        allow(subject).to receive(step).and_raise 'fishfinger in the warp core'
        allow(out).to receive(:fail)
        allow(step_out).to receive(:fail)
        allow(out).to receive(:warn)
        expect(step_out).to receive(:fail).with(match(/fishfinger in the warp core/))
        expect(out).to receive(:fail).with(match(/#{expected_message} failed, fishfinger in the warp core/))
        expect { subject.supply }.not_to raise_error
      end
    end
  end

  describe 'Running Installers' do
    context 'Installer should not be run' do
      before do
        allow(libunwind_installer).to receive(:should_install).and_return(false)
      end

      it 'does not run the installer' do
        expect(libunwind_installer).not_to receive(:install)
        subject.supply
      end

      it 'creates symbolic links' do
        expect(libunwind_installer).to receive(:create_links)
        subject.supply
      end
    end

    context 'Installer should be run' do
      it 'runs the installer' do
        allow(libunwind_installer).to receive(:should_install).and_return(true)
        expect(libunwind_installer).to receive(:install)
        subject.supply
      end

      it 'creates symbolic links' do
        expect(libunwind_installer).to receive(:create_links)
        subject.supply
      end
    end
  end

  describe 'Steps' do
    before do
      allow(subject).to receive(:should_clear_nuget_cache?).and_return(true)
    end

    describe 'FSharp runtime warning' do
      let(:warning) { "FSharp projects require runtime 1.0.x to publish" }
      before { File.write(File.join(build_dir, proj_file_name), "") }

      context 'fsharp project' do
        let(:installer) { double(:installer, descendants: [dotnet_installer]) }
        let(:dotnet_installer) do
          double(:dotnet_installer,
                 class: double(:class, name: "AspNetCoreBuildpack::DotnetSdkInstaller"),
                 version: dotnet_sdk_version,
                 cache_dir: cache_dir,
                 should_install: false,
                 create_links: nil,
                 name: "")
        end

        let(:proj_file_name) { "hello.fsproj" }

        context 'dotnet framework 1.0.x' do
          let(:dotnet_sdk_version) { '1.0.5' }

          it 'does not warn user about runtime' do
            expect(out).to_not receive(:warn).with(warning)
            subject.supply
          end
        end

        context 'dotnet framework 2.x' do
          let(:dotnet_sdk_version) { '2.0.0' }

          it 'warns user about runtime' do
            expect(out).to receive(:warn).with(warning)
            subject.supply
          end
        end
      end

      context 'csharp project' do
        let(:proj_file_name) { "hello.csproj" }
        it 'do not warn' do
          expect(out).to_not receive(:warn).with(warning)
          subject.supply
        end
      end
    end

    describe 'Restoring Cache' do
      it_behaves_like 'step', 'Restoring files from buildpack cache', :restore_cache

      context 'cache does not exist' do
        it 'skips restore' do
          expect(copier).not_to receive(:cp).with(match(cache_dir), anything, anything)
          subject.supply
        end
      end

      context 'cache exists' do
        before(:each) do
          Dir.mkdir(File.join(cache_dir, 'libunwind'))
        end

        it 'copies files from cache to build dir' do
          expect(copier).to receive(:cp).with(File.join(cache_dir, 'libunwind'), File.join(deps_dir, deps_idx), anything)
          allow(libunwind_installer).to receive(:cached?).and_return(true)
          subject.supply
        end
      end
    end

    describe 'Saving to buildpack cache' do
      it_behaves_like 'step', 'Saving to buildpack cache', :save_cache

      before(:each) do
        Dir.mkdir(File.join(deps_dir, deps_idx))
        Dir.mkdir(File.join(deps_dir, deps_idx, 'libunwind'))
      end

      it 'copies files to cache dir' do
        allow(libunwind_installer).to receive(:cached?).and_return(false)
        expect(copier).to receive(:cp).with("#{deps_dir}/#{deps_idx}/libunwind", cache_dir, anything)
        subject.send(:save_cache, out)
      end

      context 'when the files fail to copy to the cache' do
        before(:each) do
          Dir.mkdir(File.join(cache_dir, 'libunwind'))
        end

        it 'does not throw an exception' do
          allow(copier).to receive(:cp).and_raise(StandardError)
          expect(out).to receive(:fail).with(anything)
          expect { subject.send(:save_cache, out) }.not_to raise_error
        end

        it 'outputs a failure message' do
          allow(copier).to receive(:cp).and_raise(StandardError)
          expect(out).to receive(:fail).with('Failed to save cached files for libunwind')
          subject.send(:save_cache, out)
        end

        it 'removes the cache folder' do
          allow(copier).to receive(:cp).and_raise(StandardError)
          expect(out).to receive(:fail).with('Failed to save cached files for libunwind')
          subject.send(:save_cache, out)
          expect(File.exist?(File.join(cache_dir, 'libunwind'))).not_to be_truthy
        end
      end
    end
  end

end
