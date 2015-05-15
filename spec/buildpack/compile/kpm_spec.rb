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

require "rspec"
require_relative "../../../lib/buildpack.rb"

describe AspNet5Buildpack::KPM do
  let(:shell) do
    double(:shell, :env => {})
  end

  let(:out) do
    double(:out)
  end

  subject(:kpm) do
    AspNet5Buildpack::KPM.new(shell)
  end

  it "sets HOME env variable to build dir so that .kpm packages are stored in /app/.kpm" do
    allow(shell).to receive(:exec)
    kpm.restore("app-dir", out)

    expect(shell.env).to include("HOME" => "app-dir")
  end

  it "adds kpm to the PATH" do
    allow(shell).to receive(:exec)
    expect(shell).to receive(:exec).with(match("kvm install"), out)
    kpm.restore("app-dir", out)
  end

  it "sources /kvm.sh script" do
    allow(shell).to receive(:exec)
    expect(shell).to receive(:exec).with(match("bash -c 'source app-dir/.k/kvm/kvm.sh"), out)
    kpm.restore("app-dir", out)
  end

  it "restores dependencies with kpm" do
    allow(shell).to receive(:exec)
    expect(shell).to receive(:exec).with(match("kpm restore"), out)
    kpm.restore("app-dir", out)
  end
end
