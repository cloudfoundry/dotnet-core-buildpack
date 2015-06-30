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

module AspNet5Buildpack
  class Out
    def initialize(description = nil)
      puts "-----> #{description}\n" unless description.nil?
    end

    def step(description)
      Out.new(description)
    end

    def succeed
      puts "       OK\n"
    end

    def warn(message)
      puts to_warning(message)
    end

    def fail(message)
      puts "       FAILED: #{message}\n"
    end

    def print(message)
      puts "       #{message}\n"
    end

    private

    def to_warning(message)
      buff = "\n"
      buff += "  #{'*' * 72}\n"
      prefix = 'WARNING:'
      message.scan(/.{1,58}/).each do |line|
        buff += "  * #{prefix} #{line}#{' ' * (60 - line.length)}*\n"
        prefix = '        ' if prefix == 'WARNING:'
      end
      buff += "  #{'*' * 72}\n.\n"
    end
  end
end
