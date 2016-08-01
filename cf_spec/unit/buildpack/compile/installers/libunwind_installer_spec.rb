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
  subject(:installer) { described_class.new(dir, cache_dir, shell) }

  describe '#version' do
    it 'has a default version' do
      expect(subject.version).to eq('1.1')
    end
  end

  describe '#cached?' do
    context 'cache directory exists in the build directory' do
      before do
        FileUtils.mkdir_p(File.join(dir, 'libunwind'))
      end

      it 'returns true' do
        expect(installer.send(:cached?)).to be_truthy
      end
    end

    context 'cache directory does not exist in the build directory' do
      it 'returns false' do
        expect(installer.send(:cached?)).not_to be_truthy
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
      subject.install(out)
    end
  end

  describe '#should_install' do
    context 'cache folder exists' do
      it 'returns false' do
        allow(installer).to receive(:cached?).and_return(true)
        expect(installer.should_install(nil)).not_to be_truthy
      end
    end

    context 'cache folder does not exist' do
      it 'returns true' do
        allow(installer).to receive(:cached?).and_return(false)
        expect(installer.should_install(nil)).to be_truthy
      end
    end
  end
end
