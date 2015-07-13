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
require_relative '../../../lib/buildpack.rb'

describe AspNet5Buildpack::DnxInstaller do
  let(:shell) do
    double(:shell, env: {}, path: [])
  end

  let(:out) do
    double(:out)
  end

  let(:dir) do
    Dir.mktmpdir
  end

  subject(:installer) do
    AspNet5Buildpack::DnxInstaller.new(shell)
  end

  it 'sets HOME env variable to build dir so that runtimes are stored in /app/.dnx' do
    allow(shell).to receive(:exec)
    installer.install(dir, out)
    expect(shell.env).to include('HOME' => dir)
  end

  it 'adds /app/mono/bin to the path' do
    allow(shell).to receive(:exec)
    installer.install(dir, out)
    expect(shell.path).to include('/app/mono/bin')
  end

  it 'sources the dnvm script' do
    allow(shell).to receive(:exec)
    expect(shell).to receive(:exec).with(match("source #{dir}/.dnx/dnvm/dnvm.sh"), out)
    installer.install(dir, out)
  end

  describe 'dnvm installer' do
    context 'global.json does not exist' do
      it 'installs the latest version' do
        allow(shell).to receive(:exec)
        expect(shell).to receive(:exec).with(match('dnvm install latest -p -r mono'), out)
        installer.install(dir, out)
      end
    end

    context 'global.json exists' do
      before do
        json = '{ "sdk": { "version": "1.0.0-beta1" } }'
        IO.write(File.join(dir, 'global.json'), json)
      end
      it 'installs the specified version' do
        allow(shell).to receive(:exec)
        expect(shell).to receive(:exec).with(match('dnvm install 1.0.0-beta1 -p -r mono'), out)
        installer.install(dir, out)
      end
    end

    context 'global.json exists with a BOM from Visual Studio in it' do
      before do
        json = "\uFEFF{ \"sdk\": { \"version\": \"1.0.0-beta1\" } }"
        IO.write(File.join(dir, 'global.json'), json)
      end
      it 'installs the specified version' do
        allow(shell).to receive(:exec)
        expect(shell).to receive(:exec).with(match('dnvm install 1.0.0-beta1 -p -r mono'), out)
        installer.install(dir, out)
      end
    end

    context 'invalid global.json exists' do
      before do
        json = '"version": "1.0.0-beta1"'
        IO.write(File.join(dir, 'global.json'), json)
      end
      it 'warns and installs the latest version' do
        allow(shell).to receive(:exec)
        expect(shell).to receive(:exec).with(match('dnvm install latest -p -r mono'), out)
        expect(out).to receive(:warn).with("File #{dir}/global.json is not valid JSON")
        installer.install(dir, out)
      end
    end

    context 'global.json exists but does not include a version' do
      before do
        json = '{ "projects": [ "src", "test" ] }'
        IO.write(File.join(dir, 'global.json'), json)
      end
      it 'installs the latest version' do
        allow(shell).to receive(:exec)
        expect(shell).to receive(:exec).with(match('dnvm install latest -p -r mono'), out)
        installer.install(dir, out)
      end
    end
  end
end
