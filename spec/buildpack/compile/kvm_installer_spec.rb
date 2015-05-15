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

describe AspNet5Buildpack::KvmInstaller do
  let(:shell) do
    double(:shell, :env => {})
  end

  let(:out) do
    double(:out)
  end

  subject(:installer) do
    AspNet5Buildpack::KvmInstaller.new(shell)
  end

  it "runs the kvm web installer" do
    expect(shell).to receive(:exec).with("curl -s https://raw.githubusercontent.com/aspnet/Home/master/kvminstall.sh | sh", out)
    installer.install("passed-directory", out)
  end

  it "sets KRE_USER_HOME based on passed directory" do
    allow(shell).to receive(:exec)
    installer.install("passed-directory", out)

    expect(shell.env).to include("KRE_USER_HOME" => File.join("passed-directory",".k"))
  end
end
