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
require_relative 'installer'
require_relative '../scripts_parser'

module AspNetCoreBuildpack
  class NodeJsInstaller < Installer
    BOWER_COMMAND = 'bower'.freeze
    CACHE_DIR = '.node'.freeze
    NPM_COMMAND = 'npm'.freeze
    VERSION = '6.9.0'.freeze

    def cache_dir
      CACHE_DIR
    end

    def initialize(build_dir, bp_cache_dir, shell)
      @bp_cache_dir = bp_cache_dir
      @build_dir = build_dir
      @scripts_parser = ScriptsParser.new(build_dir)
      @shell = shell
    end

    def cached?
      File.exist? File.join(@bp_cache_dir, CACHE_DIR, File.basename(dependency_name, '.tar.gz'.freeze), 'bin'.freeze)
    end

    def install(out)
      dest_dir = File.join(@build_dir, CACHE_DIR)

      out.print("Node.js version: #{version}")
      @shell.exec("#{buildpack_root}/compile-extensions/bin/download_dependency #{dependency_name} /tmp", out)
      FileUtils.rm_rf(dest_dir) if File.exist?(dest_dir)
      @shell.exec("mkdir -p #{dest_dir}; tar xzf /tmp/#{dependency_name} -C #{dest_dir}", out)
    end

    def name
      'Node.js'.freeze
    end

    def path
      bin_folder if File.exist?(File.join(@build_dir, cache_dir))
    end

    def should_install(app_dir)
      published_project = app_dir.published_project
      !(published_project || cached?) && @scripts_parser.scripts_section_exists?([BOWER_COMMAND, NPM_COMMAND])
    end

    def version
      VERSION
    end

    private

    def bin_folder
      File.join('$HOME'.freeze, CACHE_DIR, File.basename(dependency_name, '.tar.gz'.freeze), 'bin'.freeze)
    end

    def dependency_name
      "node-v#{version}-linux-x64.tar.gz"
    end

    attr_reader :scripts_parser
  end
end
