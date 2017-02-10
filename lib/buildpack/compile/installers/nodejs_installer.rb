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

require_relative '../../app_dir'
require_relative '../../sdk_info'
require_relative 'installer'
require_relative '../scripts_parser'
require 'pathname'

module AspNetCoreBuildpack
  class NodeJsInstaller < Installer
    include SdkInfo
    BOWER_COMMAND = 'bower'.freeze
    CACHE_DIR = '.node'.freeze
    NPM_COMMAND = 'npm'.freeze

    def self.install_order
      2
    end

    def cache_dir
      CACHE_DIR
    end

    def initialize(build_dir, bp_cache_dir, manifest_file, shell)
      @build_dir = build_dir
      @bp_cache_dir = bp_cache_dir
      @scripts_parser = ScriptsParser.new(build_dir)
      @manifest_file = manifest_file
      @shell = shell
    end

    def cached?
      File.exist? File.join(@bp_cache_dir, CACHE_DIR, File.basename(dependency_name, '.tar.gz'.freeze), 'bin'.freeze)
    end

    def install(out)
      dest_dir = File.join(@build_dir, CACHE_DIR)

      out.print("Node.js version: #{version}")
      @shell.exec("#{buildpack_root}/compile-extensions/bin/download_dependency #{dependency_name} /tmp", out)
      @shell.exec("#{buildpack_root}/compile-extensions/bin/warn_if_newer_patch #{dependency_name} #{buildpack_root}/manifest.yml", out)
      FileUtils.rm_rf(dest_dir) if File.exist?(dest_dir)
      @shell.exec("mkdir -p #{dest_dir}; tar xzf /tmp/#{dependency_name} -C #{dest_dir}", out)
    end

    def name
      'Node.js'.freeze
    end

    def path
      "#{bin_folder}:#{node_modules_folders}" if File.exist?(File.join(@build_dir, cache_dir))
    end

    def should_install(app_dir)
      return true if ENV['INSTALL_NODE'] == 'true'

      published_project = app_dir.published_project
      !(published_project || cached?) && @scripts_parser.scripts_section_exists?([BOWER_COMMAND, NPM_COMMAND])
    end

    def version
      compile_extensions_dir = File.join(File.dirname(__FILE__), '..', '..', '..', '..', 'compile-extensions')
      @version ||= `#{compile_extensions_dir}/bin/default_version_for #{@manifest_file} node`
    end

    private

    def bin_folder
      File.join('$HOME'.freeze, CACHE_DIR, File.basename(dependency_name, '.tar.gz'.freeze), 'bin'.freeze)
    end

    def node_modules_folders
      app_dir = AppDir.new(@build_dir)
      project_dirs = app_dir.project_paths.map do |project|
        if msbuild?(@build_dir)
          File.join(@build_dir, File.dirname(project))
        else
          File.join(@build_dir, project)
        end
      end

      project_dirs.map do |dir|
        File.join('$HOME', Pathname.new(dir).relative_path_from(Pathname.new(@build_dir)).to_s, 'node_modules', '.bin')
      end.compact.join(':')
    end

    def dependency_name
      "node-v#{version}-linux-x64.tar.gz"
    end

    attr_reader :scripts_parser
  end
end
