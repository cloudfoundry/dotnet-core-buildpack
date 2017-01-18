# Encoding: utf-8
# ASP.NET Core Buildpack
# Copyright 2015-2016 the original author or authors.
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
require 'tempfile'

describe AspNetCoreBuildpack::LibunwindInstaller do
  let(:dir) { Dir.mktmpdir }
  let(:cache_dir) { Dir.mktmpdir }
  let(:shell) { AspNetCoreBuildpack::Shell.new }
  let(:out) { double(:out) }

  let(:manifest_dir)  { Dir.mktmpdir }
  let(:manifest_file) { File.join(manifest_dir, 'manifest.yml') }
  let(:manifest_contents) do
    <<-YAML
doesn't matter for these tests
    YAML
  end

  before do
    File.write(manifest_file, manifest_contents)
  end

  after do
    FileUtils.rm_rf(manifest_dir)
  end

  subject(:installer) { described_class.new(dir, cache_dir, manifest_file, shell) }

  describe '#version' do
    it 'has a default version' do
      expect(subject.version).to eq('1.2')
    end
  end

  describe '#cached?' do
    context 'cache directory exists in the buildpack cache' do
      before do
        FileUtils.mkdir_p(File.join(cache_dir, 'libunwind'))
      end

      context 'cached version is the same as the current version being installed' do
        before do
          File.open(File.join(cache_dir, 'libunwind', 'VERSION'), 'w') do |f|
            f.write '1.2'
          end
        end

        it 'returns true' do
          expect(subject.send(:cached?)).to be_truthy
        end
      end

      context 'cached version is different than the current version being installed' do
        before do
          File.open(File.join(cache_dir, 'libunwind', 'VERSION'), 'w') do |f|
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
    it 'downloads file with compile-extensions' do
      allow(shell).to receive(:exec).and_return(0)
      expect(shell).to receive(:exec) do |*args|
        cmd = args.first
        expect(cmd).to match(/download_dependency/)
        expect(cmd).to match(/tar/)
      end
      expect(out).to receive(:print).with(/libunwind version/)
      expect(subject).to receive(:write_version_file).with(anything)
      subject.install(out)
    end
  end

  describe '#should_install' do
    context 'cache folder exists' do
      it 'returns false' do
        allow(subject).to receive(:cached?).and_return(true)
        expect(subject.should_install(nil)).not_to be_truthy
      end
    end

    context 'cache folder does not exist' do
      it 'returns true' do
        allow(subject).to receive(:cached?).and_return(false)
        expect(subject.should_install(nil)).to be_truthy
      end
    end
  end
end
