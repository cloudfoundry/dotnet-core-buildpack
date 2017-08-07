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
require 'rexml/document'
require_relative '../sdk_info'
require_relative '../app_dir'

module AspNetCoreBuildpack
  class ScriptsParser
    include SdkInfo

    SCRIPTS_KEY = 'scripts'.freeze

    def initialize(build_dir, deps_dir, deps_idx)
      @build_dir = build_dir
      @build_dir_dir = deps_dir
      @deps_idx = deps_idx
      @deps_dir = deps_dir
      @app_dir = AppDir.new(@build_dir, @deps_dir, @deps_idx)
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

    def json_scripts_section_exists?(check_commands)
      return_value = false
      check_keys = ['precompile'.freeze, 'postcompile'.freeze]
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

    def xml_scripts_section_exists?(check_commands)
      @app_dir.msbuild_projects.each do |proj|
        doc = REXML::Document.new(File.read(File.join(@build_dir, proj), encoding: 'bom|utf-8'))

        targets = doc.elements.to_a('Project/Target').select do |e|
          target_valid?(e)
        end

        commands = []

        targets.each do |target|
          commands += target.elements.to_a('Exec')
        end

        commands.each do |command|
          check_commands.each do |check|
            return true if command.attributes['Command'].include? check
          end
        end
      end

      false
    end

    def scripts_section_exists?(check_commands)
      if msbuild?
        xml_scripts_section_exists?(check_commands)
      elsif project_json?
        json_scripts_section_exists?(check_commands)
      end
    end

    private

    def target_valid?(target)
      target_names = %w(BeforeBuild BeforeCompile BeforePublish AfterBuild AfterCompile AfterPublish)
      target_attributes = %w(BeforeTargets AfterTargets)

      name_matches = target_names.include? target.attributes['Name']

      attribute_matches = target_attributes.any? do |attribute|
        target.attributes.key?(attribute) && target_contains_step(target.attributes[attribute])
      end

      name_matches || attribute_matches
    end

    def target_contains_step(target)
      target_steps = %w(Build Compile Publish)

      target_steps.any? do |step|
        target.include? step
      end
    end

    attr_reader :build_dir
  end
end
