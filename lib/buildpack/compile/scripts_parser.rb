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
  class ScriptsParser
    SCRIPTS_KEY = 'scripts'.freeze

    def initialize(build_dir)
      @build_dir = build_dir
    end

    def get_scripts_section(project_json)
      project_props = JSON.parse(File.read(project_json, encoding: 'bom|utf-8'))
      project_props[SCRIPTS_KEY] if project_props.key?(SCRIPTS_KEY)
    end

    def key_array_contains_command(scripts, check_key, check_command)
      scripts2 = scripts[check_key].flat_map { |c| c.split('&&').each(&:strip!) }
      scripts2.each do |command|
        return true if command.downcase.start_with?(check_command)
      end
      false
    end

    def key_contains_command(scripts, check_key, check_command)
      return key_array_contains_command(scripts, check_key, check_command) if scripts[check_key].is_a?(Array)
      return key_string_contains_command(scripts, check_key, check_command) unless scripts[check_key].is_a?(Array)
    end

    def key_string_contains_command(scripts, check_key, check_command)
      commands = scripts[check_key].split('&&').each(&:strip!)
      commands.each do |command|
        return true if command.downcase.start_with?(check_command)
      end
      false
    end

    def scripts_section_exists?(check_commands)
      return_value = false
      check_keys = ['prebuild'.freeze, 'postbuild'.freeze, 'prerestore'.freeze, 'postrestore'.freeze]
      Dir.glob(File.join(@build_dir, '**', 'project.json'.freeze)).each do |project_json|
        scripts = get_scripts_section(project_json)
        next unless scripts
        check_keys.select { |check_key| scripts.key?(check_key) }.each do |check_key|
          check_commands.each do |check_command|
            return_value = key_contains_command(scripts, check_key, check_command)
            return return_value if return_value
          end
        end
      end
      return_value
    end

    private

    attr_reader :build_dir
  end
end
