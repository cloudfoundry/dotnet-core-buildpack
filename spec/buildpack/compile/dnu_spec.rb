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

describe AspNet5Buildpack::DNU do
  let(:shell) do
    double(:shell, env: {}, path: [])
  end

  let(:out) do
    double(:out)
  end

  subject(:dnu) do
    AspNet5Buildpack::DNU.new(shell)
  end

  it 'sets HOME env variable to build dir so that packages are stored in /app/.dnx' do
    allow(shell).to receive(:exec)
    dnu.restore('app-dir', out)

    expect(shell.env).to include('HOME' => 'app-dir')
  end

  it 'adds /app/mono/bin to the path' do
    allow(shell).to receive(:exec)
    dnu.restore('app-dir', out)
    expect(shell.path).to include('/app/mono/bin')
  end

  it 'adds dnu to the PATH' do
    allow(shell).to receive(:exec)
    expect(shell).to receive(:exec).with(match('dnvm use default'), out)
    dnu.restore('app-dir', out)
  end

  it 'sources dnvm.sh script' do
    allow(shell).to receive(:exec)
    expect(shell).to receive(:exec).with(match('source app-dir/.dnx/dnvm/dnvm.sh'), out)
    dnu.restore('app-dir', out)
  end

  it 'restores dependencies with dnu' do
    allow(shell).to receive(:exec)
    expect(shell).to receive(:exec).with(match('dnu restore'), out)
    dnu.restore('app-dir', out)
  end

end
