# Encoding: utf-8
# ASP.NET 5 Buildpack
# Copyright 2015 the original author or authors.
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

require 'json'

module AspNet5Buildpack
  class DnxInstaller

    DNX_VERSION_FILE_NAME = 'global.json'.freeze
    DEFAULT_DNX_VERSION = 'latest'.freeze

    def initialize(shell)
      @shell = shell
    end

    def install(dir, out)
      @shell.env['HOME'] = dir
      @shell.path << '/app/mono/bin'
      version = dnx_version(dir, out)
      @shell.exec("bash -c 'source #{dir}/.dnx/dnvm/dnvm.sh; dnvm install #{version} -p -r mono'", out)
    end

    private

    def dnx_version(dir, out)
      dnx_version = DEFAULT_DNX_VERSION
      version_file = File.expand_path(File.join(dir, DNX_VERSION_FILE_NAME))
      if File.exists?(version_file)
        begin
          global_props = JSON.parse(File.read(version_file, encoding: 'bom|utf-8'))
          if global_props.has_key?('sdk')
            sdk = global_props['sdk']
            dnx_version = sdk['version'] if sdk.has_key?('version')
          end
        rescue
          out.warn("File #{version_file} is not valid JSON")
        end
      end
      dnx_version
    end
  end
end
