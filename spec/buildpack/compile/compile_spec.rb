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

require "rspec"
require "tmpdir"
require "fileutils"
require_relative "../../../lib/buildpack.rb"

describe AspNet5Buildpack::Compiler do
  subject(:compiler) do
    AspNet5Buildpack::Compiler.new(buildDir, cacheDir, mono_binary, nowinDir, kvm_installer, mozroots, kre_installer, kpm, release_yml_writer, copier, out)
  end

  let(:mono_binary) do
    double(:mono_binary, :extract => nil)
  end

  let(:copier) do
    double(:copier, :cp => nil)
  end

  let(:kvm_installer) do
    double(:kvm_installer, :install => nil)
  end

  let(:kre_installer) do
    double(:kre_installer, :install => nil)
  end

  let(:kpm) do
    double(:kpm, :restore => nil)
  end

  let(:mozroots) do
    double(:mozroots, :import => nil)
  end

  let(:release_yml_writer) do
    double(:release_yml_writer, :write_release_yml => nil)
  end

  let(:out) do
    double(:out, :step => double(:unknown_step, :succeed => nil)).tap do |out|
      allow(out).to receive(:warn)
    end
  end

  let(:buildDir) do
    Dir.mktmpdir
  end

  let(:cacheDir) do
    Dir.mktmpdir
  end

  let(:nowinDir) do
    Dir.mktmpdir
  end

  it "prints a big warning message" do
    expect(out).to receive(:warn).with(match("experimental"))
    compiler.compile
  end

  shared_examples "A Step" do |expected_message, step, next_step|
    let(:step_out) do
      double(:step_out, :succeed => nil).tap do |step_out|
        allow(out).to receive(:step).with(expected_message).and_return step_out
      end
    end

    it "outputs the step name" do
      expect(out).to receive(:step).with(expected_message)
      compiler.compile
    end

    context "when it succeeds" do
      it "causes the step to succeed" do
        expect(step_out).to receive(:succeed)
        compiler.compile
      end
    end

    context "when it fails" do
      before do
        allow(subject).to receive(step).and_raise "fishfinger in the warp core"
        allow(out).to receive(:fail)
        allow(step_out).to receive(:fail)
        allow(out).to receive(:warn)
      end

      it "prints a helpful error" do
        expect(step_out).to receive(:fail).with(match(/fishfinger in the warp core/))
        expect(out).to receive(:fail).with(match(/#{expected_message} failed, fishfinger in the warp core/))
        expect(out).to receive(:warn).with(match("experimental"))
        expect { compiler.compile }.not_to raise_error
      end

      if next_step
        it "does not run further steps" do
          expect(subject).not_to receive(next_step)
          compiler.compile
        end
      end
    end
  end

  describe "Steps" do
    describe "Extracting Mono" do
      it_behaves_like "A Step", "Extracting mono", :extract_mono, :install_mozroot_certs

      it "extracts to /app" do
        expect(mono_binary).to receive(:extract).with("/app", anything)
        compiler.compile
      end

      context "when mono is already extracted because it was cached" do
        it "copies only .k to cache dir" do
          allow(File).to receive(:exist?).and_return(true)
          expect(mono_binary).not_to receive(:extract).with("/app", anything)
          compiler.compile
        end
      end
    end

    describe "Adding Nowin.vNext" do
      it_behaves_like "A Step", "Adding Nowin.vNext", :copy_nowin, :install_mozroot_certs

      it "extracts to build dir" do
        expect(copier).to receive(:cp).with(nowinDir, "#{buildDir}/src", anything)
        compiler.compile
      end
    end

    describe "Importing Certificates" do
      it_behaves_like "A Step", "Importing Mozilla Root Certificates", :install_mozroot_certs, :install_kvm

      it "imports the certificates" do
        expect(mozroots).to receive(:import)
        compiler.compile
      end
    end

    describe "Installing KVM" do
      it_behaves_like "A Step", "Installing KVM", :install_kvm, :install_kre

      it "installs kvm" do
        expect(kvm_installer).to receive(:install).with(buildDir, anything)
        compiler.compile
      end
    end

    describe "Restoring Cache" do
      it_behaves_like "A Step", "Restoring files from buildpack cache", :restore_cache, :install_kvm

      context "when the cache does not exist" do
        it "does not try copying" do
          expect(copier).not_to receive(:cp).with(match(cacheDir), anything, anything)
          compiler.compile
        end
      end

      context "when the cache exists" do
        before(:each) do
          Dir.mkdir(File.join(cacheDir, ".k"))
          Dir.mkdir(File.join(cacheDir, "mono"))
        end

        it "restores all files from the cache to build dir" do
          expect(copier).to receive(:cp).with(File.join(cacheDir, ".k"), buildDir, anything)
          expect(copier).to receive(:cp).with(File.join(cacheDir, "mono"), "/app", anything)
          compiler.compile
        end
      end
    end

    describe "Installing KRE with KVM" do
      it_behaves_like "A Step", "Installing KRE with KVM", :install_kre, :restore_dependencies

      it "installs kre" do
        expect(kre_installer).to receive(:install).with(buildDir, anything)
        compiler.compile
      end
    end

    describe "Moving files in to place" do
      it_behaves_like "A Step", "Moving files in to place", :move_to_app_dir, :save_cache

      it "copies mono to build dir" do
        expect(copier).to receive(:cp).with("/app/mono", buildDir, anything)
        compiler.compile
      end
    end

    describe "Saving to buildpack cache" do
      it_behaves_like "A Step", "Saving to buildpack cache", :save_cache, :write_release_yml

      it "copies .k and mono to cache dir" do
        expect(copier).to receive(:cp).with("#{buildDir}/.k", cacheDir, anything)
        expect(copier).to receive(:cp).with("/app/mono", cacheDir, anything)
        compiler.compile
      end

      context "when mono is already cached" do
        before(:each) do
          Dir.mkdir(File.join(cacheDir, "mono"))
        end
        it "copies only .k to cache dir" do
          expect(copier).to receive(:cp).with("#{buildDir}/.k", cacheDir, anything)
          compiler.compile
        end
      end
    end

    describe "Writing Release YML" do
      it_behaves_like "A Step", "Writing Release YML", :write_release_yml, nil

      it "writes release yml" do
        expect(release_yml_writer).to receive(:write_release_yml)
        compiler.compile
      end
    end
  end
end
