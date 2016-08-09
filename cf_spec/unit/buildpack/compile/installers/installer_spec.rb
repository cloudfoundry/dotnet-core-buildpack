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

$LOAD_PATH << 'cf_spec'
require 'spec_helper'
require 'rspec'

describe AspNetCoreBuildpack::Installer do
  let(:cache_dir) { Dir.mktmpdir }
  subject(:installer) { described_class.new }

  describe '#buildpack_root' do
    it 'returns the root directory of the buildpack' do
      rakefile = File.join(subject.send(:buildpack_root), 'Rakefile')
      expect(File.exist?(rakefile)).to be_truthy
    end
  end

  describe '#write_version_file' do
    context 'version_file is nil' do
      it 'does not raise an error' do
        expect { subject.send(:write_version_file, '1.0.0') }.not_to raise_error
      end
    end

    context 'version_file is not nil' do
      it 'writes version file' do
        allow(subject).to receive(:version_file).and_return(File.join(cache_dir, 'VERSION'))
        expect { subject.send(:write_version_file, '1.0.0') }.not_to raise_error
        expect(File.exist?(File.join(cache_dir, 'VERSION'))).to be_truthy
      end
    end
  end
end
