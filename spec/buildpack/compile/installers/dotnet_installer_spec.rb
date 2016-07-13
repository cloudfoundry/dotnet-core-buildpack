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

require 'rspec'
require_relative '../../../../lib/buildpack.rb'
require_relative '../../../../lib/buildpack/compile/installers/dotnet_installer.rb'

describe AspNetCoreBuildpack::DotnetInstaller do
  let(:dir) { Dir.mktmpdir }
  let(:cache_dir) { Dir.mktmpdir }
  let(:shell) { double(:shell, env: {}) }
  let(:out) { double(:out) }
  let(:self_contained_app_dir) { double(:self_contained_app_dir, published_project: 'project1') }
  let(:app_dir) { double(:app_dir, published_project: false, with_project_json: %w(['project1', 'project2'])) }
  subject(:installer) { AspNetCoreBuildpack::DotnetInstaller.new(dir, cache_dir, shell) }

  describe '#cached?' do
    context 'cache directory exists in the build directory' do
      before do
        FileUtils.mkdir_p(File.join(dir, '.dotnet'))
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
      expect(out).to receive(:print).with(/dotnet version/)
      subject.install(out)
    end
  end

  describe '#restore' do
    it 'runs dotnet restore' do
      expect(shell).to receive(:exec) do |*args|
        cmd = args.first
        expect(cmd).to match(/dotnet restore/)
      end
      installer.should_restore(app_dir)
      installer.restore(out)
    end
  end

  describe '#should_install' do
    context 'app is self-contained' do
      before do
        File.open(File.join(dir, 'project1'), 'w') { |f| f.write('a') }
      end

      it 'returns false' do
        expect(installer.should_install(self_contained_app_dir)).not_to be_truthy
      end
    end

    context 'app is not self-contained' do
      it 'returns true' do
        expect(installer.should_install(app_dir)).to be_truthy
      end
    end
  end

  describe '#should_restore' do
    context 'app is portable or self-contained' do
      it 'returns false' do
        expect(installer.should_restore(self_contained_app_dir)).not_to be_truthy
      end
    end

    context 'app is not portable or self-contained' do
      it 'returns true' do
        expect(installer.should_restore(app_dir)).to be_truthy
      end
    end
  end
end
