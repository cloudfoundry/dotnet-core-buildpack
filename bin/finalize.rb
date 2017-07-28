#!/usr/bin/env ruby
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
#
build_dir = ARGV[0]
cache_dir = ARGV[1]
deps_dir = ARGV[2]
deps_idx = ARGV[3]
buildpack_dir = File.join(File.dirname(__FILE__), '..')
$LOAD_PATH.unshift File.expand_path('../../lib', __FILE__)

require 'buildpack'
require 'open3'

if deps_dir
  stdout, stderr, status = Open3.capture3("#{buildpack_dir}/compile-extensions/bin/build_path_from_supply #{deps_dir}")

  if status.exitstatus.nonzero?
    puts "build_path_from_supply script failed: #{stdout} \n====\n #{stderr}"
    exit 1
  end

  stdout.split("\n").each do |line|
    var, val = line.split('=')
    ENV[var] = val
  end
end

if AspNetCoreBuildpack.finalize(build_dir, cache_dir, deps_dir, deps_idx)
  system("#{buildpack_dir}/compile-extensions/bin/store_buildpack_metadata #{buildpack_dir} #{cache_dir}")
  if deps_dir
    stdout, stderr, status = Open3.capture3("#{buildpack_dir}/compile-extensions/bin/write_profiled_from_supply #{deps_dir} #{build_dir}")
    if status.exitstatus.nonzero?
      puts "write_profiled_from_supply failed: #{stdout} \n====\n #{stderr}"
      exit 1
    end
  end
  exit 0
else
  exit 1
end
