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
    described_class.new(build_dir, cache_dir, copier, installer.descendants, out)
  end

  before do
    allow($stdout).to receive(:write)
  end

  let(:installer) { double(:installer, descendants: [libunwind_installer]) }
  let(:libunwind_installer) do
    double(:libunwind_installer, install_order: 0, install: nil).tap do |libunwind_installer|
      allow(libunwind_installer).to receive(:install_description)
      allow(libunwind_installer).to receive(:cache_dir).and_return('libunwind')
      allow(libunwind_installer).to receive(:should_install).and_return(true)
      allow(libunwind_installer).to receive(:name).and_return('libunwind')
    end
  end

  let(:copier) { double(:copier, cp: nil) }
  let(:build_dir) { Dir.mktmpdir }
  let(:cache_dir) { Dir.mktmpdir }

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
      subject.compile
    end

    it 'runs step' do
      allow(step_out).to receive(:print)
      expect(step_out).to receive(:succeed)
      allow(libunwind_installer).to receive(:cached?)
      subject.compile
    end

    context 'step fails' do
      it 'prints helpful error' do
        allow(subject).to receive(step).and_raise 'fishfinger in the warp core'
        allow(out).to receive(:fail)
        allow(step_out).to receive(:fail)
        allow(out).to receive(:warn)
        expect(step_out).to receive(:fail).with(match(/fishfinger in the warp core/))
        expect(out).to receive(:fail).with(match(/#{expected_message} failed, fishfinger in the warp core/))
        expect { subject.compile }.not_to raise_error
      end
    end
  end

  describe 'Running Installers' do
    context 'Installer should not be run' do
      it 'does not run the installer' do
        allow(libunwind_installer).to receive(:should_install).and_return(false)
        expect(libunwind_installer).not_to receive(:install)
        subject.compile
      end
    end

    context 'Installer should be run' do
      it 'runs the installer' do
        allow(libunwind_installer).to receive(:should_install).and_return(true)
        expect(libunwind_installer).to receive(:install)
        subject.compile
      end
    end
  end

  describe 'Steps' do
    before do
      allow(subject).to receive(:should_clear_nuget_cache?).and_return(true)
    end

    describe 'Restoring Cache' do
      it_behaves_like 'step', 'Restoring files from buildpack cache', :restore_cache

      context 'cache does not exist' do
        it 'skips restore' do
          expect(copier).not_to receive(:cp).with(match(cache_dir), anything, anything)
          subject.compile
        end
      end

      context 'cache exists' do
        before(:each) do
          Dir.mkdir(File.join(cache_dir, 'libunwind'))
        end

        it 'copies files from cache to build dir' do
          expect(copier).to receive(:cp).with(File.join(cache_dir, 'libunwind'), build_dir, anything)
          allow(libunwind_installer).to receive(:cached?).and_return(true)
          subject.compile
        end
      end
    end

    describe 'Clearing NuGet cache' do
      it_behaves_like 'step', 'Clearing NuGet packages cache', :clear_nuget_cache

      context 'cache exists' do
        before(:each) do
          Dir.mkdir(File.join(cache_dir, '.nuget'))
          File.open(File.join(cache_dir, '.nuget', 'Package.dll'), 'w') { |f| f.write 'test' }
        end

        it 'removes the NuGet cache folder' do
          expect(File.exist?(File.join(cache_dir, '.nuget', 'Package.dll'))).to be_truthy
          subject.compile
          expect(File.exist?(File.join(cache_dir, '.nuget', 'Package.dll'))).not_to be_truthy
        end
      end

      context 'cache does not exist' do
        it 'does not raise an exception' do
          expect { subject.compile }.not_to raise_error
        end
      end
    end

    describe 'Restoring NuGet packages cache' do
      it_behaves_like 'step', 'Restoring NuGet packages cache', :restore_nuget_cache

      context 'cache does not exist' do
        it 'skips restore' do
          expect(copier).not_to receive(:cp).with(match(cache_dir), anything, anything)
          subject.compile
        end
      end

      context 'cache exists and is valid' do
        before(:each) do
          Dir.mkdir(File.join(cache_dir, '.nuget'))
        end

        it 'copies files from cache to build dir' do
          allow(subject).to receive(:nuget_cache_is_valid?).and_return(true)
          expect(copier).to receive(:cp).with(File.join(cache_dir, '.nuget'), build_dir, anything)
          subject.compile
        end
      end

      context 'cache exists, but is not valid' do
        before(:each) do
          Dir.mkdir(File.join(cache_dir, '.nuget'))
        end

        it 'skips restoring cache' do
          allow(subject).to receive(:nuget_cache_is_valid?).and_return(false)
          expect(copier).not_to receive(:cp)
          subject.compile
        end
      end
    end

    describe 'Cleaning staging area' do
      let(:node_dir)   { File.join(build_dir, '.node') }
      let(:nuget_dir)  { File.join(build_dir, '.nuget') }
      let(:dotnet_dir) { File.join(build_dir, '.dotnet') }

      before do
        FileUtils.mkdir_p(node_dir)
        FileUtils.mkdir_p(nuget_dir)
        FileUtils.mkdir_p(dotnet_dir)
        allow(subject).to receive(:msbuild?).with(build_dir).and_return(true)
      end

      it_behaves_like 'step', 'Cleaning staging area', :clean_staging_area

      context 'project is msbuild' do
        context 'published app is self-contained' do

          before do
            publish_dir = File.join(build_dir, '.cloudfoundry', 'dotnet_publish')
            FileUtils.mkdir_p(publish_dir)
            File.write(File.join(publish_dir, 'project_name'), 'xxx')
            File.write(File.join(publish_dir, 'project_name.runtimeconfig.json'), 'xxx')
          end

          it 'removes the .dotnet, .node, and .nuget directories' do
            subject.compile
            expect(File.exist?(node_dir)).to be_falsey
            expect(File.exist?(nuget_dir)).to be_falsey
            expect(File.exist?(dotnet_dir)).to be_falsey
          end
        end

        context 'published app is portable' do
          it 'removes the .node and .nuget directories' do
            subject.compile
            expect(File.exist?(node_dir)).to be_falsey
            expect(File.exist?(nuget_dir)).to be_falsey
            expect(File.exist?(dotnet_dir)).to be_truthy
          end
        end
      end

      context 'project is project.json' do
        before do
          allow(subject).to receive(:msbuild?).with(build_dir).and_return(false)
        end

        it 'does not remove any directories' do
          subject.compile
          expect(File.exist?(node_dir)).to be_truthy
          expect(File.exist?(nuget_dir)).to be_truthy
          expect(File.exist?(dotnet_dir)).to be_truthy
        end
      end
    end

    describe 'Saving to buildpack cache' do
      it_behaves_like 'step', 'Saving to buildpack cache', :save_cache

      before(:each) do
        Dir.mkdir(File.join(build_dir, 'libunwind'))
      end

      it 'copies files to cache dir' do
        allow(libunwind_installer).to receive(:cached?).and_return(false)
        expect(copier).to receive(:cp).with("#{build_dir}/libunwind", cache_dir, anything)
        subject.send(:save_cache, out)
      end

      context 'when the cache already exists' do
        before(:each) do
          Dir.mkdir(File.join(cache_dir, 'libunwind'))
          Dir.mkdir(File.join(build_dir, '.nuget'))
        end

        it 'copies only .nuget to cache dir' do
          allow(libunwind_installer).to receive(:cached?).and_return(true)
          expect(copier).to receive(:cp).with("#{build_dir}/.nuget", cache_dir, anything)
          subject.send(:save_cache, out)
        end
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

  describe '#should_clear_nuget_cache?' do
    context 'NuGet cache exists' do
      context 'NuGet package cache is invalid' do
        before do
          allow(subject).to receive(:nuget_cache_is_valid?).and_return(false)
        end

        it 'returns true' do
          expect(subject).to receive(:should_clear_nuget_cache?).and_return(true)
          subject.compile
        end
      end

      context 'NuGet package cache is valid' do
        context 'CACHE_NUGET_PACKAGES is set to false' do
          before do
            ENV['CACHE_NUGET_PACKAGES'] = 'false'
          end

          it 'returns true' do
            expect(subject).to receive(:should_clear_nuget_cache?).and_return(true)
            subject.compile
          end
        end

        context 'CACHE_NUGET_PACKAGES is not set to false' do
          it 'returns false' do
            expect(subject).to receive(:should_clear_nuget_cache?).and_return(false)
            subject.compile
          end
        end
      end
    end

    context 'NuGet cache does not exist' do
      it 'returns false' do
        expect(subject).to receive(:should_clear_nuget_cache?).and_return(false)
        subject.compile
      end
    end
  end

  describe '#should_save_nuget_cache' do
    context '.nuget folder exists in build_dir' do
      context 'CACHE_NUGET_PACKAGES is set to false' do
        before do
          ENV['CACHE_NUGET_PACKAGES'] = 'false'
        end

        it 'returns false' do
          expect(subject).to receive(:should_save_nuget_cache?).and_return(false)
          subject.compile
        end
      end

      context 'CACHE_NUGET_PACKAGES is not set to false' do
        it 'returns true' do
          expect(subject).to receive(:should_save_nuget_cache?).and_return(false)
          subject.compile
        end
      end
    end
  end

  describe '#should_clear_nuget_cache?' do
    context 'CACHE_NUGET_PACKAGES is set to false' do
      before do
        ENV['CACHE_NUGET_PACKAGES'] = 'false'
      end

      context 'cache folder exists' do
        before do
          FileUtils.mkdir_p(File.join(cache_dir, '.nuget'))
        end

        it 'returns true' do
          expect(subject.send(:should_clear_nuget_cache?)).to be_truthy
        end
      end

      context 'cache folder does not exist' do
        it 'returns false' do
          expect(subject.send(:should_clear_nuget_cache?)).not_to be_truthy
        end
      end
    end

    context 'CACHE_NUGET_PACKAGES is not set to false' do
      it 'returns false' do
        expect(subject.send(:should_clear_nuget_cache?)).not_to be_truthy
      end
    end
  end
end
