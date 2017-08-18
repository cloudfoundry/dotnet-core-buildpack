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
require_relative '../app_dir'

module AspNetCoreBuildpack
  class DotnetSdkVersion
    def initialize(build_dir, deps_dir, deps_idx, manifest_file)
      @manifest_file = manifest_file
      @build_dir = build_dir
      @deps_dir = deps_dir
      @deps_idx = deps_idx
      @global_json_file_name = 'global.json'
      @sdk_tools_file = File.join(File.dirname(manifest_file), 'dotnet-sdk-tools.yml')
      @out = Out.new
      @app_dir = AppDir.new(@build_dir, @deps_dir, @deps_idx)
      @dotnet_sdk_tooling = ENV['DOTNET_SDK_TOOLING']
    end

    def version
      sdk_version = sdk_version_to_install

      deprecation_warning = "Support for project.json in the .NET Core buildpack will\n" \
                            "be deprecated. For more information see:\n" \
                            'https://blogs.msdn.microsoft.com/dotnet/2016/11/16/announcing-net-core-tools-msbuild-alpha'

      out.warn(deprecation_warning) if project_json_sdk_versions.include? sdk_version
      sdk_version
    end

    private

    def fsharp_project?
      Dir.glob(File.join(@build_dir, '**', '*.fsproj')).any?
    end

    def default_sdk_version
      return @default_sdk_version if @default_sdk_version
      if fsharp_project?
        @default_sdk_version = '1.0.4'
      else
        buildpack_root = File.join(File.dirname(__FILE__), '..', '..', '..')
        @default_sdk_version = `#{buildpack_root}/compile-extensions/bin/default_version_for #{@manifest_file} dotnet`
      end
    end

    def sdk_version_to_install
      sdk_version = version_from_global_json
      sdk_version = convert_missing_sdk_version_to_default(sdk_version)
      return sdk_version if sdk_version

      app_has_project_json = @app_dir.with_project_json.any?
      app_has_msbuild_projects = @app_dir.msbuild_projects.any?

      if app_has_msbuild_projects && app_has_project_json
        warning = "Found both project.json and MSBuild projects in app:\n" \
                  "MSBuild projects: #{@app_dir.msbuild_projects.join(', ')}\n" \
                  "Directories with project.json: #{@app_dir.with_project_json.join(', ')}\n" \
                  "Please provide a global.json file that specifies the\n" \
                  'correct .NET SDK version for this app'

        out.warn(warning)

        raise 'App contains both project.json and MSBuild projects'
      elsif app_has_project_json
        project_json_sdk_versions.last
      else
        default_sdk_version
      end
    end

    def convert_missing_sdk_version_to_default(sdk_version)
      if sdk_version && !sdk_version_exists(sdk_version)
        warning = "SDK #{sdk_version} not available,\nusing the default SDK version(#{default_sdk_version})"
        out.warn(warning)
        return default_sdk_version
      end
      sdk_version
    end

    def msbuild_sdk_versions
      @msbuild_sdk_versions ||= YAML.load_file(@sdk_tools_file)['msbuild']
    end

    def project_json_sdk_versions
      @project_sdk_versions ||= YAML.load_file(@sdk_tools_file)['project_json']
    end

    def version_from_global_json
      global_json_file = File.expand_path(File.join(@build_dir, @global_json_file_name))
      return nil unless File.exist?(global_json_file)
      begin
        global_props = JSON.parse(File.read(global_json_file, encoding: 'bom|utf-8'))
        if global_props.key?('sdk')
          sdk = global_props['sdk']
          return sdk['version'] if sdk.key?('version')
        end
      rescue
        out.warn("File #{global_json_file} is not valid JSON")
      end
      nil
    end

    def sdk_version_exists(version)
      msbuild_sdk_versions.include?(version) || project_json_sdk_versions.include?(version)
    end

    attr_reader :out
  end
end
