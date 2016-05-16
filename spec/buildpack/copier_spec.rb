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

require 'rspec'
require 'tmpdir'
require_relative '../../lib/buildpack.rb'

describe AspNetCoreBuildpack::Copier do
  let(:src) { Dir.mktmpdir }
  let(:dest) { Dir.mktmpdir }
  let(:out) { double(:out, print: nil) }
  let!(:dir1) { File.join(src, 'dir1').tap { |d| Dir.mkdir(d) } }

  let!(:file1) do
    File.join(src, 'file1').tap do |f|
      File.open(f, 'w') { |w| w.write('something') }
    end
  end

  let!(:one_level_deep) do
    File.join(dir1, 'one_level_deep').tap do |f|
      File.open(f, 'w') { |w| w.write('something') }
    end
  end

  describe '#cp' do
    it 'copies all files from source to dest' do
      subject.cp(src, dest, out)
      expect(Dir[File.join(dest, '**/*')]).to include(
        File.join(dest, File.basename(src), File.basename(file1)),
        File.join(dest, File.basename(src), 'dir1', File.basename(one_level_deep)))
    end

    it 'prints the number of files copied and the src/dest' do
      expect(out).to receive(:print).with("Copied 4 files from #{src} to #{dest}")
      subject.cp(src, dest, out)
    end
  end
end
