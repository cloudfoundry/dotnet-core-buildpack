# Encoding: utf-8
# ASP.NET Core Buildpack
# Copyright 2016 the original author or authors.
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
  class DotnetInstaller
    def initialize(shell)
      @shell = shell
    end

    def install(dir, out)
      @shell.env['HOME'] = dir
      install_script_url = 'https://raw.githubusercontent.com/dotnet/cli/rel/1.0.0-preview1/scripts/obtain/dotnet-install.sh'
      cmd = "bash -c 'DOTNET_INSTALL_SKIP_PREREQS=1 source <(curl -sSL #{install_script_url})'"
      @shell.exec(cmd, out)
    end

    def should_install(dir)
      published_project = AppDir.new(dir).published_project
      if published_project && File.exist?(File.join(dir, published_project))
        return false
      end
      true
    end
  end
end
