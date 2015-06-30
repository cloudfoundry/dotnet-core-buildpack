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
require_relative '../../../lib/buildpack.rb'

describe AspNet5Buildpack::DnvmInstaller do
  let(:shell) do
    double(:shell, env: {})
  end

  let(:out) do
    double(:out)
  end

  subject(:installer) do
    AspNet5Buildpack::DnvmInstaller.new(shell)
  end

  it 'creates .bashrc so dnvminstall.sh does not complain' do
    expect(shell).to receive(:exec).with(match('touch ~/.bashrc'), out)
    installer.install('passed-directory', out)
  end

  it 'runs the dnvm web installer' do
    cmd = 'curl -sSL https://raw.githubusercontent.com/aspnet/Home/dev/dnvminstall.sh | DNX_BRANCH=dev sh'
    expect(shell).to receive(:exec).with(match(cmd), out)
    installer.install('passed-directory', out)
  end

  it 'deletes .bashrc because dnvminstall.sh updated it with temporary paths' do
    expect(shell).to receive(:exec).with(match('rm -rf ~/.bashrc'), out)
    installer.install('passed-directory', out)
  end

  it 'sets HOME based on passed directory' do
    allow(shell).to receive(:exec)
    installer.install('passed-directory', out)

    expect(shell.env).to include('HOME' => 'passed-directory')
  end
end
