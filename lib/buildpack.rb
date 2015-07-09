# Encoding: utf-8
# ASP.NET 5 Buildpack
# Copyright 2014-2015 the original author or authors.
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

require_relative './buildpack/compiler.rb'
require_relative './buildpack/detecter.rb'
require_relative './buildpack/shell.rb'
require_relative './buildpack/out.rb'
require_relative './buildpack/copier.rb'

module AspNet5Buildpack
  def self.detect(build_dir)
    Detecter.new.detect(build_dir)
  end

  def self.compile(build_dir, cache_dir)
    compiler(build_dir, cache_dir).compile
  end

  def self.compiler(build_dir, cache_dir)
    Compiler.new(
      build_dir,
      cache_dir,
      MonoInstaller.new(build_dir, shell),
      File.expand_path('../../resources/Nowin.vNext', __FILE__),
      DnvmInstaller.new(shell),
      Mozroots.new(shell),
      DnxInstaller.new(shell),
      DNU.new(shell),
      ReleaseYmlWriter.new,
      Copier.new,
      out)
  end

  def self.out
    @out ||= Out.new
  end

  def self.shell
    @shell ||= Shell.new
  end

end
