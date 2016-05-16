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

module AspNetCoreBuildpack
  class GetTextInstaller
    VERSION = '0.19.7'.freeze

    def initialize(app_dir, shell)
      @app_dir = app_dir
      @shell = shell
    end

    def extract(dest_dir, out)
      out.print("gettext version: #{version}")
      cmd = "mkdir -p #{dest_dir}; curl -L `translate_dependency_url #{dependency_name}` -s | tar zxv -C #{dest_dir} &> /dev/null"
      run_common_cmd(cmd, out)
    end

    def gettext_tar_gz(out)
      run_common_cmd("translate_dependency_url #{dependency_name}", out)
    end

    def version
      VERSION
    end

    private

    def run_common_cmd(cmd, out)
      commoncmd = "bash -c 'export BUILDPACK_PATH=#{buildpack_root}; source $BUILDPACK_PATH/compile-extensions/lib/common &> /dev/null; #{cmd}'"
      @shell.exec(commoncmd, out)
    end

    def buildpack_root
      current_dir = File.expand_path(File.dirname(__FILE__))
      File.dirname(File.dirname(File.dirname(current_dir)))
    end

    def dependency_name
      "gettext-x-#{version}.tar.gz"
    end
  end
end
