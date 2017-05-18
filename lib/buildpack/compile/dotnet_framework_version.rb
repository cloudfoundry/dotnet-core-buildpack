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

require 'yaml'
require 'json'
require 'rexml/document'
require_relative '../sdk_info'
require_relative '../app_dir'

module AspNetCoreBuildpack
  class DotnetFrameworkVersion
    include SdkInfo

    def initialize(build_dir, nuget_cache_dir)
      @build_dir = build_dir
      @nuget_cache_dir = nuget_cache_dir
      @out = Out.new
    end

    def versions
      runtime_config_json_file = Dir.glob(File.join(@build_dir, '*.runtimeconfig.json')).first

      framework_versions = []

      if !runtime_config_json_file.nil?
        runtime_config_framework = get_version_from_runtime_config_json(runtime_config_json_file)
        framework_versions.push runtime_config_framework unless runtime_config_framework.nil?
      elsif restored_framework_versions.any?
        out.print("Detected .NET Core runtime version(s) #{needed_framework_versions.join(', ')} required according to 'dotnet restore'")
        framework_versions += needed_framework_versions
      else
        raise 'Unable to determine .NET Core runtime version(s) to install'
      end

      framework_versions.uniq
    end

    private

    def needed_framework_versions
      version_hash = {}

      restored_framework_versions.each do |ver|
        major, minor, = ver.split('.')
        version_line = "#{major}.#{minor}"

        if version_hash[version_line].nil?
          version_hash[version_line] = [ver]
        else
          version_hash[version_line].push ver
        end
      end

      required_versions = version_hash.values.map do |v|
        v.sort { |a, b| Gem::Version.new(a) <=> Gem::Version.new(b) }.last
      end

      required_versions += runtime_framework_versions if msbuild?(@build_dir)

      required_versions.sort { |a, b| Gem::Version.new(a) <=> Gem::Version.new(b) }
    end

    def runtime_framework_versions
      AppDir.new(@build_dir).msbuild_projects.map do |proj|
        doc = REXML::Document.new(File.read(File.join(@build_dir, proj), encoding: 'bom|utf-8'))

        runtime_version = doc.elements.to_a('Project/PropertyGroup/RuntimeFrameworkVersion').first

        runtime_version.text unless runtime_version.nil?
      end.compact
    end

    def restored_framework_versions
      if project_json?(@build_dir)
        netcore_app_dir = 'Microsoft.NETCore.App'
      elsif msbuild?(@build_dir)
        netcore_app_dir = 'microsoft.netcore.app'
      end

      Dir.glob(File.join(@nuget_cache_dir, 'packages', netcore_app_dir, '*')).sort.map do |path|
        File.basename(path)
      end
    end

    def get_version_from_runtime_config_json(runtime_config_json_file)
      begin
        global_props = JSON.parse(File.read(runtime_config_json_file, encoding: 'bom|utf-8'))
      rescue
        raise "#{runtime_config_json_file} contains invalid JSON"
      end

      has_framework_version = global_props.key?('runtimeOptions') &&
                              global_props['runtimeOptions'].key?('framework') &&
                              global_props['runtimeOptions']['framework'].key?('version')

      return nil unless has_framework_version

      version = global_props['runtimeOptions']['framework']['version']
      out.print("Detected .NET Core runtime version #{version} in #{runtime_config_json_file}")

      version
    end

    attr_reader :out
  end
end
