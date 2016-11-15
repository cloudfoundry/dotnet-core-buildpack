$LOAD_PATH.push "#{File.dirname(__FILE__)}/../../vendor/iniparse-1.4.2/lib"

require 'iniparse'

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
    DEPLOYMENT_FILE_NAME = '.deployment'.freeze

    def initialize(dir)
      @dir = dir
    end

    def root
      @dir
    end

    def with_command(cmd)
      with_project_json.select { |d| !commands(d)[cmd].nil? && commands(d)[cmd] != '' }
    end

    def with_project_json
      Dir.glob(File.join(@dir, '**', 'project.json')).map do |d|
        Pathname.new(File.dirname(d)).relative_path_from(Pathname.new(@dir))
      end
    end

    def project_json(dir)
      File.join(@dir, dir, 'project.json')
    end

    def commands(dir)
      JSON.parse(IO.read(project_json(dir), encoding: 'bom|utf-8')).fetch('commands', {})
    end

    def deployment_file_project
      project_path = nil

      paths = with_project_json
      deployment_file = File.expand_path(File.join(@dir, DEPLOYMENT_FILE_NAME))

      if File.exist?(deployment_file)
        deployment_ini = IniParse.parse(File.read(deployment_file, encoding: 'bom|utf-8'))
        deployment_project_path = deployment_ini['config']['project']

        if deployment_project_path.nil?
          raise DeploymentConfigError.new("must have project key")
        elsif deployment_project_path.class == Array
          raise DeploymentConfigError.new("must only contain one project key")
        end

        project_suffix = deployment_project_path.split('.').last

        if project_suffix == 'xproj' || project_suffix == 'csproj'
          path = Pathname.new(File.dirname(deployment_project_path))
        else
          path = Pathname.new(deployment_project_path)
        end

        project_path = path if paths.include?(path)
      end

      project_path
    end

    def main_project_path
      path = deployment_file_project
      project_paths = with_project_json
      multiple_paths = project_paths.any? && !project_paths.one?
      raise 'Multiple paths contain a project.json file, but no .deployment file was used' unless path || !multiple_paths
      path = project_paths.first unless path
      path if path
    end

    def published_project
      config_files = Dir.glob(File.join(@dir, '*.runtimeconfig.json'))
      m = /(.*)[.]runtimeconfig[.]json/i.match(Pathname.new(config_files.first).basename.to_s) if config_files.one?
      m[1].to_s unless m.nil?
    end
  end
end
