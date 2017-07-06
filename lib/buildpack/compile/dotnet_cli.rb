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

require_relative '../sdk_info'
require_relative '../app_dir'

module AspNetCoreBuildpack
  class DotnetCli
    include SdkInfo
    PUBLISH_DIR = File.join('.cloudfoundry', 'dotnet_publish')

    def initialize(build_dir, installers)
      @build_dir = build_dir
      @installers = installers
      @app_dir = AppDir.new(@build_dir)
      @shell = AspNetCoreBuildpack.shell
    end

    def restore(out)
      setup_shell_environment

      if msbuild?(@build_dir)
        @app_dir.project_paths.each do |project|
          cmd = "bash -c 'cd #{@build_dir}; dotnet restore #{project}'"
          @shell.exec(cmd, out)
        end
      else
        project_list = @app_dir.project_paths.join(' ')
        cmd = "bash -c 'cd #{@build_dir}; dotnet restore #{project_list}'"
        @shell.exec(cmd, out)
      end
    end

    def publish(out)
      setup_shell_environment

      main_project = @app_dir.main_project_path
      raise 'No project found to build' if main_project.nil?

      publish_dir = File.join(@build_dir, PUBLISH_DIR)
      FileUtils.mkdir_p(publish_dir)

      cmd = "bash -c 'cd #{@build_dir}; dotnet publish #{main_project} -o #{publish_dir} -c #{publish_config}'"

      @shell.exec(cmd, out)
    end

    private

    def publish_config
      if ENV['PUBLISH_RELEASE_CONFIG'] == 'true'
        'Release'
      else
        'Debug'
      end
    end

    def setup_shell_environment
      project_dirs = @app_dir.project_paths.map do |project|
        if msbuild?(@build_dir)
          File.join(@build_dir, File.dirname(project))
        else
          File.join(@build_dir, project)
        end
      end

      node_modules_paths = project_dirs.map do |dir|
        File.join(dir, 'node_modules', '.bin')
      end.compact.join(':')

      if msbuild?(@build_dir)
        @shell.env['DOTNET_SKIP_FIRST_TIME_EXPERIENCE'] = "true"
      end

      @shell.env['HOME'] = @build_dir
      @shell.env['LD_LIBRARY_PATH'] = "$LD_LIBRARY_PATH:#{@build_dir}/libunwind/lib"
      @shell.env['PATH'] = "$PATH:#{@installers.map(&:path).compact.join(':')}:#{node_modules_paths}"
    end
  end
end
