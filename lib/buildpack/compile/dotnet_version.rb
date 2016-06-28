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

require 'json'

module AspNetCoreBuildpack
  class DotnetVersion
    DOTNET_VERSION_FILE_NAME = 'global.json'.freeze
    DEFAULT_DOTNET_VERSION = '1.0.0-preview2-003121'.freeze

    def version(dir, out)
      dotnet_version = DEFAULT_DOTNET_VERSION
      version_file = File.expand_path(File.join(dir, DOTNET_VERSION_FILE_NAME))
      if File.exist?(version_file)
        begin
          global_props = JSON.parse(File.read(version_file, encoding: 'bom|utf-8'))
          if global_props.key?('sdk')
            sdk = global_props['sdk']
            dotnet_version = sdk['version'] if sdk.key?('version')
          end
        rescue
          out.warn("File #{version_file} is not valid JSON")
        end
      end
      dotnet_version
    end
  end
end
