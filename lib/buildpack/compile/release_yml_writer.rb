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

require 'fileutils'
require 'json'
require_relative 'dnx_version'

module AspNet5Buildpack
  class ReleaseYmlWriter
    def write_release_yml(build_dir, out)
      dirs = Dirs.new(build_dir, out)
      write_startup_script(dirs.startup_script_path)
      write_yml(dirs, out)
    end

    private

    def write_yml(dirs, out)
      unless dirs.with_existing_cfweb.empty?
        write_yml_for(dirs.release_yml_path, dirs.with_existing_cfweb.first, 'cf-web')
        return
      end

      path = dirs.main_project_path
      if path
        version = DnxVersion.new.version(dirs.root, out)
        add_cfweb_command(dirs.project_json(path), version)
        write_yml_for(dirs.release_yml_path, path, 'cf-web')
        return
      end

      write_yml_for(dirs.release_yml_path, '.', 'cf-web')
    end

    def add_cfweb_command(project_json, version)
      json = JSON.parse(IO.read(project_json, encoding: 'bom|utf-8'))
      json['dependencies'] ||= {}
      json['dependencies']['Kestrel'] = version unless json['dependencies']['Kestrel']
      json['commands'] ||= {}
      json['commands']['cf-web'] = 'Microsoft.AspNet.Hosting --server Kestrel'
      IO.write(project_json, JSON.pretty_generate(json))
    end

    def write_startup_script(path)
      FileUtils.mkdir_p(File.dirname(path))
      File.open(path, 'w') do |f|
        f.write 'export HOME=/app;'
        f.write 'export PATH=$HOME/mono/bin:$PATH;'
        f.write 'export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/libuv/lib;'
        f.write 'source $HOME/.dnx/dnvm/dnvm.sh;'
        f.write 'dnvm use default -r mono -a x64;'
        f.write 'dnu restore;'
      end
    end

    def write_yml_for(ymlPath, web_dir, cmd)
      File.open(ymlPath, 'w') do |f|
        f.write <<EOT
---
default_process_types:
  web: cd #{web_dir}; sleep 999999 | dnx . #{cmd} --server.urls http://${VCAP_APP_HOST}:${PORT}
EOT
      end
    end

    class Dirs
      DEPLOYMENT_FILE_NAME = '.deployment'.freeze

      def initialize(dir, out)
        @dir = dir
        @out = out
      end

      def root
        @dir
      end

      def release_yml_path
        File.join(@dir, 'aspnet5-buildpack-release.yml')
      end

      def startup_script_path
        File.join(@dir, '.profile.d', 'startup.sh')
      end

      def with_existing_cfweb
        with_command('cf-web')
      end

      def with_command(cmd)
        with_project_json.select { |d| !commands(d)[cmd].nil? && commands(d)[cmd] != '' }
      end

      def with_project_json
        Dir.glob(File.join(@dir, '**', 'project.json')).map do |d|
          relative_path_to(File.dirname(d))
        end
      end

      def commands(dir)
        JSON.load(IO.read(project_json(dir), encoding: 'bom|utf-8')).fetch('commands', {})
      end

      def project_json(dir)
        File.join(@dir, dir, 'project.json')
      end

      def relative_path_to(d)
        Pathname.new(d).relative_path_from(Pathname.new(@dir))
      end

      def main_project_path
        paths = with_project_json
        deployment_file = File.expand_path(File.join(@dir, DEPLOYMENT_FILE_NAME))
        if File.exist?(deployment_file)
          File.foreach(deployment_file) do |line|
            m = /project = (.*)/.match(line)
            if m
              path = Pathname.new(m[1])
              return path if paths.include?(path)
            end
          end
        end
        paths.first
      end
    end
  end
end
