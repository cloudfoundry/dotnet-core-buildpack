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

module AspNetCoreBuildpack
  class Detecter
    def detect(dir)
      fromsource = false

      dirs_to_check = Dir.glob(File.join(dir, '**', 'project.json')).map { |file| File.dirname(file) }

      proj_files_regex = /.+\.(?:csproj|fsproj|vbproj)/
      dirs_to_check += Dir.glob(File.join(dir, '**', '*.??proj')).grep(proj_files_regex).map { |file| File.dirname(file) }

      dirs_to_check.each do |directory|
        fromsource = Dir.glob(File.join(directory, '**', '*.??')).grep(/.+\.(?:cs|fs|vb)/).any?
        break if fromsource
      end
      frompublish = Dir.glob(File.join(dir, '*.runtimeconfig.json')).any?
      fromsource || frompublish
    end
  end
end
