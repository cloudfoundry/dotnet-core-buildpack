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

require 'yaml'
require 'json'

module AspNetCoreBuildpack
  class DotnetSdkVersion
    def initialize(build_dir, manifest_file)
      buildpack_root = File.join(File.dirname(__FILE__), '..', '..', '..')

      @build_dir = build_dir
      @global_json_file_name = 'global.json'
      @default_sdk_version = `#{buildpack_root}/compile-extensions/bin/default_version_for #{manifest_file} dotnet`
      @out = Out.new
    end

    def version
      sdk_version = @default_sdk_version
      global_json_file = File.expand_path(File.join(@build_dir, @global_json_file_name))

      if File.exist?(global_json_file)
        sdk_version = get_version_from_global_json(global_json_file)
      end

      sdk_version
    end

    private

    def get_version_from_global_json(global_json_file)
      begin
        global_props = JSON.parse(File.read(global_json_file, encoding: 'bom|utf-8'))
        if global_props.key?('sdk')
          sdk = global_props['sdk']
          return sdk['version'] if sdk.key?('version')
        end
      rescue
        out.warn("File #{global_json_file} is not valid JSON")
      end
      @default_sdk_version
    end

    attr_reader :out
  end
end
