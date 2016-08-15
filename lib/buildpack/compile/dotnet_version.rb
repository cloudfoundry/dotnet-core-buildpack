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
    GLOBAL_JSON_FILE_NAME = 'global.json'.freeze
    DEFAULT_DOTNET_VERSION = '1.0.0-preview2-003121'.freeze
    DOTNET_RUNTIME_VERSIONS = { '1.0.0-rc2-3002702'.freeze => '1.0.0-preview1-002702'.freeze,
                                '1.0.0'.freeze => '1.0.0-preview2-003121'.freeze }.freeze

    def version(dir, out)
      dotnet_version = DEFAULT_DOTNET_VERSION
      global_json_file = File.expand_path(File.join(dir, GLOBAL_JSON_FILE_NAME))
      runtimeconfig_json_file = Dir.glob(File.join(dir, '*.runtimeconfig.json')).first
      if File.exist?(global_json_file)
        dotnet_version = get_version_from_global_json(global_json_file, out)
      elsif !runtimeconfig_json_file.nil?
        dotnet_version = get_version_from_runtime_config_json(runtimeconfig_json_file, out)
      end
      dotnet_version
    end

    private

    def get_version_from_global_json(global_json_file, out)
      begin
        global_props = JSON.parse(File.read(global_json_file, encoding: 'bom|utf-8'))
        if global_props.key?('sdk')
          sdk = global_props['sdk']
          return sdk['version'] if sdk.key?('version')
        end
      rescue
        out.warn("File #{global_json_file} is not valid JSON")
      end
      DEFAULT_DOTNET_VERSION
    end

    def get_version_from_runtime_config_json(runtime_config_json_file, out)
      begin
        global_props = JSON.parse(File.read(runtime_config_json_file, encoding: 'bom|utf-8'))
        if global_props.key?('runtimeOptions') && global_props['runtimeOptions'].key?('framework')
          framework = global_props['runtimeOptions']['framework']
          return DOTNET_RUNTIME_VERSIONS[framework['version']] if framework.key?('version') && DOTNET_RUNTIME_VERSIONS.key?(framework['version'])
        end
      rescue
        out.warn("File #{runtime_config_json_file} is not valid JSON")
      end
      DEFAULT_DOTNET_VERSION
    end
  end
end
