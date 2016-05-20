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

require 'rspec'
require 'tmpdir'
require_relative '../../../lib/buildpack.rb'

describe AspNetCoreBuildpack::DotnetVersion do
  let(:out) { double(:out) }
  let(:dir) { Dir.mktmpdir }

  describe '#version' do
    context 'global.json does not exist' do
      it 'resolves to the latest version' do
        expect(subject.version(dir, out)).to eq('latest')
      end
    end

    context 'global.json exists' do
      before do
        json = '{ "sdk": { "version": "1.0.0-beta1" } }'
        IO.write(File.join(dir, 'global.json'), json)
      end

      it 'resolves to the specified version' do
        expect(subject.version(dir, out)).to eq('1.0.0-beta1')
      end
    end

    context 'global.json exists with a BOM from Visual Studio in it' do
      before do
        json = "\uFEFF{ \"sdk\": { \"version\": \"1.0.0-beta1\" } }"
        IO.write(File.join(dir, 'global.json'), json)
      end

      it 'resolves to the specified version' do
        expect(subject.version(dir, out)).to eq('1.0.0-beta1')
      end
    end

    context 'invalid global.json exists' do
      before do
        json = '"version": "1.0.0-beta1"'
        IO.write(File.join(dir, 'global.json'), json)
      end

      it 'warns and resolves to the latest version' do
        expect(out).to receive(:warn).with("File #{dir}/global.json is not valid JSON")
        expect(subject.version(dir, out)).to eq('latest')
      end
    end

    context 'global.json exists but does not include a version' do
      before do
        json = '{ "projects": [ "src", "test" ] }'
        IO.write(File.join(dir, 'global.json'), json)
      end

      it 'resolves to the latest version' do
        expect(subject.version(dir, out)).to eq('latest')
      end
    end
  end
end
