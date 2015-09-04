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
  class DnxVersion
    DNX_VERSION_FILE_NAME = 'global.json'.freeze
    DEFAULT_DNX_VERSION = 'latest'.freeze

    def version(dir, out)
      dnx_version = DEFAULT_DNX_VERSION
      version_file = File.expand_path(File.join(dir, DNX_VERSION_FILE_NAME))
      if File.exist?(version_file)
        begin
          global_props = JSON.parse(File.read(version_file, encoding: 'bom|utf-8'))
          if global_props.key?('sdk')
            sdk = global_props['sdk']
            dnx_version = sdk['version'] if sdk.key?('version')
          end
        rescue
          out.warn("File #{version_file} is not valid JSON")
        end
      end
      dnx_version
    end
  end
end
