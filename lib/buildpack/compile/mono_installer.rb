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

module AspNet5Buildpack
  class MonoInstaller

    # really 4.0.1.44 fix release - needed for mozroots error
    DEFAULT_VERSION = '4.0.1'

    def initialize(app_dir, shell)
      @app_dir = app_dir
      @shell = shell
    end

    def extract(dest_dir, out)
      out.print("Mono version: #{version}")
      cmd = "mkdir -p #{dest_dir}; curl -L `translate_dependency_url #{dependency_name}` -s | tar zxv -C #{dest_dir} &> /dev/null"
      run_common_cmd(cmd, out)
    end

    def version
      desired_version_from_monoversion_file || DEFAULT_VERSION
    end

    def mono_tar_gz(out)
      run_common_cmd("translate_dependency_url #{dependency_name}", out)
    end

    private

    def run_common_cmd(cmd, out)
      commoncmd = "bash -c 'export BUILDPACK_PATH=#{buildpack_root}; source $BUILDPACK_PATH/compile-extensions/lib/common &> /dev/null; #{cmd}'"
      shell.exec(commoncmd, out)
    end

    def buildpack_root
      current_dir = File.expand_path(File.dirname(__FILE__))
      File.dirname(File.dirname(File.dirname(current_dir)))
    end

    def dependency_name
      "mono-x-#{version}.tar.gz"
    end

    def desired_version_from_monoversion_file
      IO.read(mono_version_file) if File.exists?(mono_version_file)
    end

    def mono_version_file
      File.join(app_dir, '.mono-version')
    end

    private

    attr_reader :app_dir
    attr_reader :shell
  end
end
