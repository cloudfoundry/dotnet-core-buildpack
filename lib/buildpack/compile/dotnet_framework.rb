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

require_relative '../out'
require_relative 'dotnet_framework_version'
require 'fileutils'

module AspNetCoreBuildpack
  class DotnetFramework
    def initialize(build_dir, nuget_cache_dir, dotnet_install_dir, shell)
      @build_dir = build_dir
      @nuget_cache_dir = nuget_cache_dir
      @dotnet_install_dir = dotnet_install_dir
      @shell = shell
    end

    def install(out)
      buildpack_root = File.join(File.dirname(__FILE__), '..', '..', '..')

      versions.each do |version|
        if installed?(version)
          out.print(".NET Core runtime #{version} already installed")
          next
        end

        out.print("Downloading and installing .NET Core runtime #{version}")
        @shell.exec("#{buildpack_root}/compile-extensions/bin/download_dependency #{dependency_name(version)} /tmp", out)
        @shell.exec("mkdir -p #{@dotnet_install_dir}; tar xzf /tmp/#{dependency_name(version)} -C #{@dotnet_install_dir}", out)
      end

      clean_up_frameworks(out)
    end

    def name
      '.NET Core runtime'.freeze
    end

    private

    def clean_up_frameworks(out)
      frameworks_to_remove = installed_frameworks - versions

      frameworks_to_remove.each do |version|
        out.print("Removing .NET Core runtime #{version}")
        FileUtils.rm_rf(File.join(@dotnet_install_dir, 'shared', 'Microsoft.NETCore.App', version))
      end
    end

    def installed_frameworks
      Dir.glob(File.join(@dotnet_install_dir, 'shared', 'Microsoft.NETCore.App', '*')).map do |path|
        File.basename(path)
      end
    end

    def installed?(version)
      installed_frameworks.include? version
    end

    def dependency_name(version)
      "dotnet-framework.#{version}.linux-amd64.tar.gz"
    end

    def versions
      @versions ||= DotnetFrameworkVersion.new(@build_dir, @nuget_cache_dir).versions
    end
  end
end
