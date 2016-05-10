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

module AspNetCoreBuildpack
  class Releaser
    KESTREL_CMD = 'kestrel'.freeze
    WEB_CMD = 'web'.freeze

    def release(build_dir)
      app = AppDir.new(build_dir)
      path = main_project_path(app)
      raise "No #{KESTREL_CMD} or #{WEB_CMD} command found" unless path
      cfweb_cmd = get_cfweb_cmd(app, path)
      write_startup_script(startup_script_path(build_dir))
      generate_yml(cfweb_cmd, build_dir, path)
    end

    private

    def write_startup_script(startup_script)
      FileUtils.mkdir_p(File.dirname(startup_script))
      File.open(startup_script, 'w') do |f|
        f.write 'export HOME=/app;'
        f.write 'export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/libuv/lib:$HOME/libunwind/lib;'
        f.write '[ -f $HOME/.dnx/dnvm/dnvm.sh ] && { source $HOME/.dnx/dnvm/dnvm.sh; dnvm use default; }'
      end
    end

    def generate_yml(cfweb_cmd, base_dir, web_dir)
      start_cmd = File.exist?(File.join(base_dir, 'approot', cfweb_cmd)) ? "approot/#{cfweb_cmd}" : "dnx --project #{web_dir} #{cfweb_cmd}"
      yml = <<-EOT
---
default_process_types:
  web: #{start_cmd} --server.urls http://0.0.0.0:${PORT}
EOT
      yml
    end

    def main_project_path(app)
      path = app.deployment_file_project
      return path if path
      kestrel_paths = app.with_project_json.select { |p| cfweb_path_exists(app, p) }
      kestrel_paths.first
    end

    def startup_script_path(dir)
      File.join(dir, '.profile.d', 'startup.sh')
    end

    def get_cfweb_cmd(app, path)
      return KESTREL_CMD if app.commands(path)[KESTREL_CMD]
      WEB_CMD
    end

    def cfweb_path_exists(app, path)
      app.commands(path)[KESTREL_CMD] || app.commands(path)[WEB_CMD]
    end
  end
end
