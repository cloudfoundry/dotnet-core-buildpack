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
require_relative '../../lib/buildpack.rb'

describe AspNet5Buildpack::Shell do
  let(:out) do
    double(:out)
  end

  it 'executes a command and returns the output' do
    expect(out).to receive(:print).with('foo')
    subject.exec('echo foo', out)
  end

  it 'executes a command and returns the stderr output' do
    expect(out).to receive(:print).with('foo')
    subject.exec('echo foo 1>&2', out)
  end

  it 'raises an exception containing the exit code if the command fails' do
    expect { subject.exec('exit 12', out) }.to raise_error(/12/)
  end

  context 'setting environment variables' do
    it 'appends an environment variable to future calls' do
      expect(out).to receive(:print).with('BAR')

      subject.env['FOO'] = 'BAR'
      subject.exec('echo $FOO', out)
    end

    it 'adds to the process path' do
      expect(out).to receive(:print).with(match(/mono/))
      subject.path << 'mono'
      subject.exec('echo $PATH', out)
    end
  end
end
