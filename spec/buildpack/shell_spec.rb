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
require_relative '../../lib/buildpack.rb'

describe AspNetCoreBuildpack::Shell do
  let(:out) { double(:out) }

  describe '#exec' do
    context 'command succeeds' do
      it 'prints stdout' do
        expect(out).to receive(:print).with('foo')
        subject.exec('echo foo', out)
      end

      it 'prints stderr' do
        expect(out).to receive(:print).with('foo')
        subject.exec('echo foo 1>&2', out)
      end
    end

    context 'command fails' do
      it 'raises an exception' do
        expect { subject.exec('exit 12', out) }.to raise_error(/12/)
      end
    end

    context 'environment variable set' do
      it 'command uses environment variable' do
        expect(out).to receive(:print).with('BAR')
        subject.env['FOO'] = 'BAR'
        subject.exec('echo $FOO', out)
      end
    end

    context 'PATH set' do
      it 'command uses PATH' do
        expect(out).to receive(:print).with(match(/mono/))
        subject.path << 'mono'
        subject.exec('echo $PATH', out)
      end
    end
  end
end
