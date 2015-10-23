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

require 'rspec'
require_relative '../../lib/buildpack.rb'

describe AspNet5Buildpack::Out do
  describe '#step' do
    it 'prints step title prefixed with arrow' do
      expect($stdout).to receive(:puts).with("-----> foo\n")
      subject.step('foo')
    end
  end

  describe '#warn' do
    it 'prints warning message surrounded asterisks' do
      expect($stdout).to receive(:puts).with("\n" \
       "  ************************************************************************\n" \
       "  * WARNING: xyz abc 123 should wrap blah blah blah foo bar baz bing bo  *\n" \
       "  *          o. this is the first message of line 2.                     *\n" \
       "  ************************************************************************\n" \
       ".\n")
      subject.warn('xyz abc 123 should wrap blah blah blah foo bar baz bing boo. this is the first message of line 2.')
    end
  end

  describe '#fail' do
    it "prints indented failure message prefixed with 'FAILED'" do
      expect($stdout).to receive(:puts).with("       FAILED: foo\n")
      subject.fail('foo')
    end
  end

  describe '#succeed' do
    it 'prints indednted OK' do
      expect($stdout).to receive(:puts).with("       OK\n")
      subject.succeed
    end
  end

  describe '#print' do
    it 'prints indented message' do
      expect($stdout).to receive(:puts).with("       foo\n")
      subject.print('foo')
    end
  end
end
