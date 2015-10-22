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

  it 'sources the dnvm script' do
    allow(shell).to receive(:exec)
    expect(shell).to receive(:exec).with(match("source #{dir}/.dnx/dnvm/dnvm.sh"), out)
    installer.install(dir, out)
  end

  it 'installs DNX' do
    allow(shell).to receive(:exec)
    expect(shell).to receive(:exec).with(match('dnvm install latest -p -r coreclr'), out)
    installer.install(dir, out)
  end
end
