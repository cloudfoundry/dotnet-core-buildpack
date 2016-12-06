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
  class Releaser
    include SdkInfo

    def release(build_dir)
      @build_dir = build_dir
      app = AppDir.new(build_dir)
      start_cmd = get_start_cmd(app)

      raise 'No project could be identified to run' if start_cmd.nil? || start_cmd.empty?

      write_startup_script(startup_script_path(build_dir))
      generate_yml(start_cmd)
    end

    private

    def write_startup_script(startup_script)
      FileUtils.mkdir_p(File.dirname(startup_script))
      File.open(startup_script, 'w') do |f|
        f.write 'export HOME=/app;'
        f.write 'export NugetPackageRoot=/app/.nuget/packages/;' if msbuild?(@build_dir)
        installers = AspNetCoreBuildpack::Installer.descendants

        library_path = get_library_path(installers)
        custom_library_path = ENV['LD_LIBRARY_PATH']
        library_path = "#{library_path}:#{custom_library_path}" if custom_library_path
        f.write "export LD_LIBRARY_PATH=#{library_path};"

        binary_path = get_binary_path(installers)
        f.write "export PATH=#{binary_path};"
      end
    end

    def generate_yml(start_cmd)
      yml = <<-EOT
---
default_process_types:
  web: #{start_cmd} --server.urls http://0.0.0.0:${PORT}
EOT
      yml
    end

    def get_binary_path(installers)
      bin_paths = installers.map do |subclass|
        subclass.new(@build_dir, @cache_dir, manifest_file, @shell).path
      end
      bin_paths.insert(0, '$PATH')
      bin_paths.compact.join(':')
    end

    def get_library_path(installers)
      library_paths = installers.map do |subclass|
        subclass.new(@build_dir, @cache_dir, manifest_file, @shell).library_path
      end
      library_paths.insert(0, '$LD_LIBRARY_PATH')
      library_paths.insert(1, '$HOME/ld_library_path')
      library_paths.compact.join(':')
    end

    def get_source_start_cmd(project)
      verbosity = ENV['BP_DEBUG'].nil? ? '' : '--verbose '
      return "dotnet #{verbosity}run --project #{project}" unless project.nil?
    end

    def get_published_start_cmd(project, build_dir)
      return "./#{project}" if !project.nil? && File.exist?(File.join(build_dir, project.to_s))
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
  end
end
