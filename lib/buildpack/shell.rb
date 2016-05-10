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

require 'open3'

module AspNetCoreBuildpack
  class Shell
    def exec(cmd, out)
      Open3.popen2e(expand(cmd)) do |_, oe, t|
        oe.each do |line|
          out.print line.chomp
        end

        raise "command failed, exit status #{t.value.exitstatus}" unless t.value.success?
      end
    end

    def path
      @path ||= []
    end

    def env
      @env ||= {}
    end

    private

    def expand(cmd)
      (exports + [cmd]).join(';')
    end

    def exports
      env.map { |k, v| "export #{k}=#{v}" } + ["export PATH=$PATH:#{path.join(':')}"]
    end
  end
end
