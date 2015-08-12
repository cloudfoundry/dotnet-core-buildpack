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

require_relative 'app_dir'

module AspNet5Buildpack
  class ReleaseYmlWriter
    CFWEB_CMD = 'kestrel'.freeze

    def write_release_yml(build_dir, out)
      dirs = AppDir.new(build_dir, out)
      path = main_project_path(dirs)
      fail 'No application found' unless path
      fail "No #{CFWEB_CMD} command found in #{path}" unless dirs.commands(path)[CFWEB_CMD]
      write_startup_script(dirs)
      write_yml(dirs.release_yml_path, path)
    end

    private

    def write_startup_script(dirs)
      FileUtils.mkdir_p(File.dirname(dirs.startup_script_path))
      File.open(dirs.startup_script_path, 'w') do |f|
        f.write 'export HOME=/app;'
        f.write 'export PATH=$HOME/mono/bin:$PATH;'
        f.write 'export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/libuv/lib;'
        f.write 'source $HOME/.dnx/dnvm/dnvm.sh;'
        f.write 'dnvm use default -r mono -a x64;'
      end
    end

    def write_yml(ymlPath, web_dir)
      File.open(ymlPath, 'w') do |f|
        f.write <<EOT
---
default_process_types:
  web: cd #{web_dir}; sleep 999999 | dnx . #{CFWEB_CMD} --server.urls http://${VCAP_APP_HOST}:${PORT}
EOT
      end
    end

    def main_project_path(dirs)
      path = dirs.deployment_file_project
      return path if path
      dirs.with_project_json.sort { |p| dirs.commands(p)[CFWEB_CMD] ? 0 : 1 }.first
    end
  end
end
