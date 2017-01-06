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
require_relative '../../out'
require_relative '../dotnet_sdk_version'
require_relative 'installer'

module AspNetCoreBuildpack
  class DotnetSdkInstaller < Installer
    include SdkInfo

    CACHE_DIR = '.dotnet'.freeze

    def self.install_order
      1
    end

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
      dest_dir = File.join(@build_dir, CACHE_DIR)

      out.print(".NET SDK version: #{version}")
      @shell.exec("#{buildpack_root}/compile-extensions/bin/download_dependency #{dependency_name} /tmp", out)
      @shell.exec("#{buildpack_root}/compile-extensions/bin/warn_if_newer_patch #{dependency_name} #{buildpack_root}/manifest.yml", out)
      @shell.exec("mkdir -p #{dest_dir}; tar xzf /tmp/#{dependency_name} -C #{dest_dir}", out)
      write_version_file(version)
    end

    def name
      '.NET SDK'.freeze
    end

    def path
      bin_folder if File.exist?(File.join(@build_dir, cache_dir))
    end

    def should_install(app_dir)
      !self_contained_project?(app_dir) && !cached?
    end

    def self_contained_project?(app_dir)
      published_project = app_dir.published_project
      published_project && File.exist?(File.join(@build_dir, published_project))
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

    def version
      @version ||= DotnetSdkVersion.new(@build_dir, @manifest_file).version
    end

    attr_reader :app_dir
    attr_accessor :dotnet_restored
  end
end
