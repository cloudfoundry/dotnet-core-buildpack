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

require_relative '../app_dir'
require_relative '../sdk_info'

module AspNetCoreBuildpack
  class StartCommandWriter
    include SdkInfo

    def initialize(build_dir, deps_dir, deps_idx)
      @build_dir = build_dir
      @deps_dir = deps_dir
      @deps_idx = deps_idx
    end

    def run
      app_root_dir = if File.exist?(File.join(@build_dir, DotnetCli::PUBLISH_DIR))
                       DotnetCli::PUBLISH_DIR
                     else
                       '.'
                     end

      app = AppDir.new(File.expand_path(File.join(@build_dir, app_root_dir)), @deps_dir, @deps_idx)
      start_cmd = get_start_cmd(app)

      raise 'No project could be identified to run' if start_cmd.nil? || start_cmd.empty?

      write_startup_script(startup_script_path(@build_dir), start_cmd)
      generate_yml(start_cmd, app_root_dir)
    end

    private

    def write_startup_script(startup_script, start_cmd)
      FileUtils.mkdir_p(File.dirname(startup_script))
      File.open(startup_script, 'w') do |f|
        f.write 'export HOME=/app;'
        f.write 'export ASPNETCORE_URLS=http://0.0.0.0:${PORT};'
        f.write "export PID=`ps -C '#{start_cmd}' -o pid= | tr -d '[:space:]'`"
      end
    end

    def generate_yml(start_cmd, app_root_dir)
      yml = <<-EOT
---
default_process_types:
  web: cd #{app_root_dir} && #{start_cmd} --server.urls http://0.0.0.0:${PORT}
EOT
      yml
    end

    def get_source_start_cmd(project)
      verbosity = ENV['BP_DEBUG'].nil? ? '' : '--verbose '
      return "dotnet #{verbosity}run --project #{project}" unless project.nil?
    end

    def get_published_start_cmd(project, build_dir)
      if !project.nil? && File.exist?(File.join(build_dir, project.to_s))
        FileUtils.chmod '+x', File.join(build_dir, project.to_s)
        return "./#{project}"
      end
      return "dotnet #{project}.dll" if File.exist? File.join(build_dir, "#{project}.dll")
      nil
    end

    def get_start_cmd(app)
      start_cmd = get_source_start_cmd(app.main_project_path)
      return start_cmd unless start_cmd.nil?

      start_cmd = get_published_start_cmd(app.published_project, app.root)
      return start_cmd unless start_cmd.nil?
    end

    def startup_script_path(dir)
      File.join(dir, '.profile.d', 'startup.sh')
    end

    def manifest_file
      File.join(File.dirname(__FILE__), '..', '..', '..', 'manifest.yml')
    end

    attr_reader :build_dir
    attr_reader :deps_dir
    attr_reader :deps_idx
  end
end
