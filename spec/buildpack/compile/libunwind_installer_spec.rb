# Encoding: utf-8
# ASP.NET 5 Buildpack
# Copyright 2015 the original author or authors.
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
require 'tempfile'
require_relative '../../../lib/buildpack.rb'
require_relative '../../../lib/buildpack/shell.rb'

describe AspNet5Buildpack::LibunwindInstaller do
  let(:dir) do
    Dir.mktmpdir
  end

  let(:shell) do
    AspNet5Buildpack::Shell.new
  end

  let(:out) do
    double(:out)
  end

  subject(:libunwind_installer) do
    described_class.new(dir, shell)
  end

  describe 'libunwind version' do
    it 'uses default version' do
      expect(subject.version).to eq('1.1')
    end
  end

  describe 'libunwind file location' do
    context 'when present in dependencies dir' do
      it 'extracts the local binary' do
        begin
          dependencies = File.expand_path(File.join(File.dirname(__FILE__), '..', '..', '..', 'dependencies'))
          FileUtils.mkdir_p dependencies
          expect(out).to receive(:print).with(%r{file:///})
          subject.libunwind_tar_gz(out)
        ensure
          FileUtils.rm_rf(dependencies) if File.exist? dependencies
        end
      end
    end

    context 'when not present in dependencies dir' do
      it 'downloads and extracts the binary' do
        expect(out).to receive(:print).with(%r{https://})
        subject.libunwind_tar_gz(out)
      end
    end
  end

  describe 'libunwind extraction' do
    it 'uses compile-extensions' do
      allow(shell).to receive(:exec).and_return(0)
      expect(shell).to receive(:exec) do |*args|
        cmd = args.first
        expect(cmd).to match(/curl/)
        expect(cmd).to match(/translate_dependency_url/)
        expect(cmd).to match(/tar/)
      end
      expect(out).to receive(:print).with(/libunwind version/)
      subject.extract(dir, out)
    end
  end
end
