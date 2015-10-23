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

require_relative '../app_dir'

module AspNet5Buildpack
  class Releaser
    CFWEB_CMD = 'kestrel'.freeze

    def release(build_dir)
      app = AppDir.new(build_dir)
      path = main_project_path(app)
      fail 'No application found' unless path
      fail "No #{CFWEB_CMD} command found in #{path}" unless app.commands(path)[CFWEB_CMD]
      write_startup_script(startup_script_path(build_dir))
      generate_yml(build_dir, path)
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

    def generate_yml(base_dir, web_dir)
      start_cmd = File.exist?(File.join(base_dir, 'approot', CFWEB_CMD)) ? "approot/#{CFWEB_CMD}" : "dnx --project #{web_dir} #{CFWEB_CMD}"
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
      app.with_project_json.sort { |p| app.commands(p)[CFWEB_CMD] ? 0 : 1 }.first
    end

    def startup_script_path(dir)
      File.join(dir, '.profile.d', 'startup.sh')
    end
  end
end
