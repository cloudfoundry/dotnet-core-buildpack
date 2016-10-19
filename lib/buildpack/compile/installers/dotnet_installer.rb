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
require_relative '../dotnet_version'
require_relative 'installer'

module AspNetCoreBuildpack
  class DotnetInstaller < Installer
    CACHE_DIR = '.dotnet'.freeze

    def cache_dir
      CACHE_DIR
    end

    def initialize(build_dir, bp_cache_dir, manifest_file, shell)
      @bp_cache_dir = bp_cache_dir
      @build_dir = build_dir
      @manifest_file = manifest_file
      @shell = shell
    end

    def cached?
      # File.open can't create the directory structure
      return false unless File.exist? File.join(@bp_cache_dir, CACHE_DIR)
      cached_version = File.open(cached_version_file, File::RDONLY | File::CREAT).select { |line| line.chomp == version }
      !cached_version.empty?
    end

    def install(out)
      buildpack_root = File.join(File.dirname(__FILE__), '..', '..', '..', '..')
      dotnet_versions_file = File.join(buildpack_root, 'dotnet-versions.yml')

      @version = DotnetVersion.new(@build_dir, @manifest_file, dotnet_versions_file, out).version

      dest_dir = File.join(@build_dir, CACHE_DIR)

      out.print("dotnet version: #{version}")
      @shell.exec("#{buildpack_root}/compile-extensions/bin/download_dependency #{dependency_name} /tmp", out)
      @shell.exec("mkdir -p #{dest_dir}; tar xzf /tmp/#{dependency_name} -C #{dest_dir}", out)
      write_version_file(@version)
    end

    def name
      'Dotnet CLI'.freeze
    end

    def path
      bin_folder if File.exist?(File.join(@build_dir, cache_dir))
    end

    def restore(out)
      @shell.env['HOME'] = @build_dir
      @shell.env['LD_LIBRARY_PATH'] = "$LD_LIBRARY_PATH:#{build_dir}/libunwind/lib"
      @shell.env['PATH'] = "$PATH:#{path}"
      project_list = @app_dir.with_project_json.join(' ')

      verbosity = ENV['BP_DEBUG'].nil? ? 'minimal' : 'debug'
      cmd = "bash -c 'cd #{build_dir}; dotnet restore --verbosity #{verbosity} #{project_list}'"
      @shell.exec(cmd, out)
    end

    def should_install(app_dir)
      published_project = app_dir.published_project
      no_install = published_project && File.exist?(File.join(@build_dir, published_project))
      !(no_install || cached?)
    end

    def should_restore(app_dir)
      @app_dir = app_dir
      published_project = app_dir.published_project
      !published_project
    end

    private

    def bin_folder
      File.join('$HOME'.freeze, CACHE_DIR)
    end

    def cache_folder
      File.join(bp_cache_dir, CACHE_DIR)
    end

    def dependency_name
      "dotnet.#{version}.linux-amd64.tar.gz"
    end

    attr_reader :app_dir
    attr_reader :version
  end
end
