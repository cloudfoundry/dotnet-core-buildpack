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

require_relative './buildpack/compile/compiler.rb'
require_relative './buildpack/compile/finalizer.rb'
require_relative './buildpack/compile/start_command_writer.rb'
require_relative './buildpack/detect/detecter.rb'
require_relative './buildpack/shell.rb'
require_relative './buildpack/out.rb'
require_relative './buildpack/copier.rb'

module AspNetCoreBuildpack
  def self.detect(build_dir)
    Detecter.new.detect(build_dir)
  end

  def self.supply(build_dir, cache_dir, deps_dir, deps_idx)
    supplier(build_dir, cache_dir, deps_dir, deps_idx).supply
  end

  def self.supplier(build_dir, cache_dir, deps_dir, deps_idx)
    manifest_file = File.join(File.dirname(__FILE__), '..', 'manifest.yml')

    Compiler.new(
      build_dir,
      cache_dir,
      deps_dir,
      deps_idx.to_s,
      Copier.new,
      [AspNetCoreBuildpack::LibunwindInstaller, AspNetCoreBuildpack::DotnetSdkInstaller, AspNetCoreBuildpack::NodeJsInstaller].map { |b| b.new(build_dir, cache_dir, deps_dir, deps_idx, manifest_file, shell) },
      out
    )
  end

  def self.finalize(build_dir, cache_dir, deps_dir, deps_idx)
    finalizer(build_dir, cache_dir, deps_dir, deps_idx).finalize
  end

  def self.finalizer(build_dir, cache_dir, deps_dir, deps_idx)
    manifest_file = File.join(File.dirname(__FILE__), '..', 'manifest.yml')

    Finalizer.new(
      build_dir,
      cache_dir,
      deps_dir,
      deps_idx.to_s,
      Copier.new,
      [AspNetCoreBuildpack::BowerInstaller, AspNetCoreBuildpack::DotnetSdkInstaller].map { |b| b.new(build_dir, cache_dir, deps_dir, deps_idx, manifest_file, shell) },
      out
    )
  end

  def self.out
    @out ||= Out.new
  end

  def self.shell
    @shell ||= Shell.new
  end
end
