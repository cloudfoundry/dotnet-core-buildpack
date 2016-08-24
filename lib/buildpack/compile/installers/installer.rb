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

module AspNetCoreBuildpack
  class Installer
    VERSION_FILE = 'VERSION'.freeze

    def self.descendants
      ObjectSpace.each_object(Class).select { |subclass| subclass < self }
    end

    def self.install_order
      1
    end

    def cached?
      false
    end

    def cache_dir
      nil
    end

    def install_description
      'Installing'
    end

    def library_path
      nil
    end

    def path
      nil
    end

    def should_install(_app_dir)
      false
    end

    protected

    def buildpack_root
      current_dir = File.expand_path(File.dirname(__FILE__))
      File.dirname(File.dirname(File.dirname(File.dirname(current_dir))))
    end

    def cached_version_file
      File.join(@bp_cache_dir, cache_dir, VERSION_FILE) unless cache_dir.nil? || @bp_cache_dir.nil?
    end

    def version_file
      File.join(@build_dir, cache_dir, VERSION_FILE) unless cache_dir.nil? || @build_dir.nil?
    end

    def write_version_file(version)
      File.open(version_file, 'w') do |f|
        f.write(version)
      end unless version_file.nil?
    end

    attr_reader :build_dir
    attr_reader :bp_cache_dir
    attr_reader :shell
  end
end
