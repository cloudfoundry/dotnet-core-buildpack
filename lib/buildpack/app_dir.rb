$LOAD_PATH.push "#{File.dirname(__FILE__)}/../../vendor/iniparse-1.4.2/lib"

require 'iniparse'
require_relative 'sdk_info'

# Encoding: utf-8

# ASP.NET Core Buildpack
# Copyright 2015-2016 the original author or authors.
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

module AspNetCoreBuildpack
  class DeploymentConfigError < StandardError
    def initialize(error_reason)
      super("Invalid .deployment file: #{error_reason}")
    end
  end

  class AppDir
    include SdkInfo

    DEPLOYMENT_FILE_NAME = '.deployment'.freeze

    def initialize(build_dir, deps_dir, deps_idx)
      @build_dir = build_dir
      @deps_dir = deps_dir
      @deps_idx = deps_idx
    end

    def root
      @build_dir
    end

    def with_project_json
      Dir.glob(File.join(@build_dir, '**', 'project.json')).map do |d|
        Pathname.new(File.dirname(d)).relative_path_from(Pathname.new(@build_dir))
      end
    end

    def msbuild_projects
      cs_projects = Dir.glob(File.join(@build_dir, '**', '*.csproj')).map do |d|
        Pathname.new(d).relative_path_from(Pathname.new(@build_dir))
      end
      fs_projects = Dir.glob(File.join(@build_dir, '**', '*.fsproj')).map do |d|
        Pathname.new(d).relative_path_from(Pathname.new(@build_dir))
      end
      vb_projects = Dir.glob(File.join(@build_dir, '**', '*.vbproj')).map do |d|
        Pathname.new(d).relative_path_from(Pathname.new(@build_dir))
      end

      cs_projects + fs_projects + vb_projects
    end

    def project_paths
      if msbuild?
        msbuild_projects
      elsif project_json?
        with_project_json
      else
        []
      end
    end

    def deployment_file_project
      project_path = nil
      paths = project_paths
      deployment_file = File.expand_path(File.join(@build_dir, DEPLOYMENT_FILE_NAME))

      if File.exist?(deployment_file)
        deployment_ini = IniParse.parse(File.read(deployment_file, encoding: 'bom|utf-8'))
        deployment_project = deployment_ini['config']['project']

        raise DeploymentConfigError, 'must have project key' if deployment_project.nil?
        raise DeploymentConfigError, 'must only contain one project key' if deployment_project.class == Array

        path = project_json? ? get_project_dir(deployment_project) : Pathname.new(deployment_project)
        path = path_without_dot_slash_prefix(path)

        project_path = path if paths.include?(path)
      end

      project_path
    end

    def get_project_dir(project_path)
      project_suffix = project_path.split('.').last

      if project_suffix == 'xproj' || project_suffix == 'csproj'
        Pathname.new(File.dirname(project_path))
      else
        Pathname.new(project_path)
      end
    end

    def main_project_path
      path = deployment_file_project
      found_project_paths = project_paths
      multiple_paths = found_project_paths.any? && !found_project_paths.one?
      raise 'Multiple paths contain a project.json file, but no .deployment file was used' unless path || !multiple_paths
      path = found_project_paths.first unless path
      path if path
    end

    def published_project
      config_files = Dir.glob(File.join(@build_dir, '*.runtimeconfig.json'))
      m = /(.*)[.]runtimeconfig[.]json/i.match(Pathname.new(config_files.first).basename.to_s) if config_files.one?
      return m[1].to_s if m

      config_files = Dir.glob(File.join(@build_dir, '.cloudfoundry', 'dotnet_publish', '*.runtimeconfig.json'))
      m = /(.*)[.]runtimeconfig[.]json/i.match(Pathname.new(config_files.first).basename.to_s) if config_files.one?
      return m[1].to_s if m
    end

    private

    def path_without_dot_slash_prefix(path)
      path_string = path.to_s
      if path_string.start_with?('./')
        Pathname.new(path_string[2..-1])
      else
        path
      end
    end
  end
end
