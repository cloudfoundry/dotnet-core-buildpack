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

describe AspNetCoreBuildpack::DotnetSdkInstaller do
  let(:dir) { Dir.mktmpdir }
  let(:cache_dir) { Dir.mktmpdir }
  let(:deps_dir) { Dir.mktmpdir }
  let(:deps_idx) { '10' }
  let(:shell) { double(:shell, env: {}) }
  let(:out) { double(:out) }
  let(:self_contained_app_dir) { double(:self_contained_app_dir, published_project: 'project1') }
  let(:app_dir) { double(:app_dir, published_project: false) }
  let(:manifest_dir)  { Dir.mktmpdir }
  let(:manifest_file) { File.join(manifest_dir, 'manifest.yml') }
  let(:manifest_contents) do
    <<-YAML
doesn't matter for these tests
    YAML
  end

  before do
    allow(AspNetCoreBuildpack::DotnetSdkVersion).to receive(:new).with(any_args).and_return(double(version: '4.4.4-002222'))

    File.write(manifest_file, manifest_contents)
  end

  after do
    FileUtils.rm_rf(manifest_dir)
    FileUtils.rm_rf(dir)
  end

  subject(:installer) { described_class.new(dir, cache_dir, deps_dir, deps_idx, manifest_file, shell) }

  describe '#version' do
    it 'is always defined' do
      expect(installer.send(:version)).to_not eq(nil)
    end
  end

  describe '#cached?' do
    context 'cache directory exists in the buildpack cache' do
      before do
        FileUtils.mkdir_p(File.join(cache_dir, 'dotnet'))
      end

      context 'cached version is the same as the current version being installed' do
        before do
          File.open(File.join(cache_dir, 'dotnet', 'VERSION'), 'w') do |f|
            f.write '1.0.0-preview2-003121'
          end
        end

        it 'returns true' do
          allow(subject).to receive(:version).and_return('1.0.0-preview2-003121')
          expect(subject.send(:cached?)).to be_truthy
        end
      end

      context 'cached version is different than the current version being installed' do
        before do
          File.open(File.join(cache_dir, 'dotnet', 'VERSION'), 'w') do |f|
            f.write '1.0.0-preview2-003131'
          end
        end

        it 'returns false' do
          expect(subject.send(:cached?)).not_to be_truthy
        end
      end
    end

    context 'cache directory does not exist in the build directory' do
      it 'returns false' do
        expect(subject.send(:cached?)).not_to be_truthy
      end
    end
  end

  describe '#install' do
    it 'downloads file with compile-extensions and writes a version file' do
      allow(shell).to receive(:exec).and_return(0)
      expect(shell).to receive(:exec) do |*args|
        cmd = args.first
        expect(cmd).to match(/download_dependency/)
        expect(cmd).to match(/4.4.4-002222/)
        expect(cmd).to match(/tar/)
      end
      expect(out).to receive(:print).with(/.NET SDK version: /)
      expect(subject).to receive(:write_version_file).with(anything)
      subject.install(out)
    end
  end

  describe '#create_links' do
    it 'creates necessary links to deps_dir/deps_idx/bin' do
      allow(shell).to receive(:exec).and_return(0)
      expect(shell).to receive(:exec) do |*args|
        cmd = args.first
        expect(cmd).to match(/cd #{File.join(deps_dir,deps_idx,'bin')}/)
        expect(cmd).to match(/ln -s ..\/dotnet\/dotnet dotnet/)
      end
      subject.create_links(out)
    end
  end

  describe '#should_install' do
    context 'app is self-contained' do
      before do
        File.open(File.join(dir, 'project1'), 'w') { |f| f.write('a') }
      end

      it 'returns false' do
        expect(subject.should_install(self_contained_app_dir)).not_to be_truthy
      end
    end

    context 'app is not self-contained' do
      it 'returns true' do
        expect(subject.should_install(app_dir)).to be_truthy
      end
    end
  end

  describe '#should_restore' do
    context 'app is portable or self-contained' do
      it 'returns false' do
        expect(subject.should_restore(self_contained_app_dir)).not_to be_truthy
      end
    end

    context 'app is not portable or self-contained' do
      it 'returns true' do
        expect(subject.should_restore(app_dir)).to be_truthy
      end
    end
  end

  describe '#write_version_file' do
    before do
      FileUtils.mkdir_p(File.join(deps_dir, deps_idx, 'dotnet'))
    end

    it 'writes a version file with the current .NET version' do
      subject.send(:write_version_file, '1.0.0')
      expect(File.exist?(File.join(deps_dir, deps_idx, 'dotnet', 'VERSION'))).to be_truthy
    end
  end
end
