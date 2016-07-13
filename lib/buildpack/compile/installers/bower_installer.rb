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
  class BowerInstaller < Installer
    BOWER_COMMAND = 'bower'.freeze
    VERSION = '1.7.9'.freeze

    def self.install_order
      2
    end

    def initialize(build_dir, bp_cache_dir, shell)
      @bp_cache_dir = bp_cache_dir
      @build_dir = build_dir
      @scripts_parser = ScriptsParser.new(build_dir)
      @shell = shell
    end

    def install(out)
      # get latest npm version path
      npm_path = Dir.glob(File.join(@build_dir, '.node', '*', 'bin')).last
      # fail if NPM is not installed
      fail 'Could not find NPM' if npm_path.nil?

      out.print("Bower version: #{version}")
      @shell.exec("#{buildpack_root}/compile-extensions/bin/download_dependency #{dependency_name} /tmp", out)
      @shell.exec("PATH=$PATH:#{npm_path} npm install -g /tmp/#{dependency_name}", out)
    end

    def install_description
      'Installing Bower'.freeze
    end

    def should_install(app_dir)
      published_project = app_dir.published_project
      !(published_project || cached?) && @scripts_parser.scripts_section_exists?([BOWER_COMMAND])
    end

    def version
      VERSION
    end

    private

    def buildpack_root
      current_dir = File.expand_path(File.dirname(__FILE__))
      File.dirname(File.dirname(File.dirname(File.dirname(current_dir))))
    end

    def cached?
      npm_path = Dir.glob(File.join(@build_dir, '.node', '*', 'bin')).last
      bower_path = File.join(npm_path, 'bower') unless npm_path.nil?
      File.exist? bower_path unless bower_path.nil?
    end

    def dependency_name
      "bower-#{version}.tgz"
    end
  end
end
