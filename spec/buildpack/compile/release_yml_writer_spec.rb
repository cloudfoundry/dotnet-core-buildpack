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
require "yaml"
require "tmpdir"
require "fileutils"
require_relative "../../../lib/buildpack.rb"

describe AspNet5Buildpack::ReleaseYmlWriter do
  let(:buildDir) do
    Dir.mktmpdir
  end

  let(:out) do
    double(:out)
  end

  describe "the release yml" do
    let(:yml_path) do
      File.join(buildDir, "aspnet5-buildpack-release.yml")
    end

    let(:yml) do
      subject.write_release_yml(buildDir, out)
      YAML.load_file(yml_path)
    end

    let(:profile_d_script) do
      subject.write_release_yml(buildDir, out)
      IO.read(File.join(buildDir, ".profile.d", "startup.sh"))
    end

    describe "the .profile.d script" do
      let(:web_dir) do
        File.join(buildDir, "foo").tap { |f| Dir.mkdir(f) }
      end

      it "should add /app/mono/bin to the path" do
        expect(profile_d_script).to include("export PATH=/app/mono/bin:$PATH;")
      end

      it "should set HOME to /app (so that dependencies are picked up from /app/.kre etc" do
        expect(profile_d_script).to include("export HOME=/app")
      end

      it "should source kvm script" do
        expect(profile_d_script).to include("source /app/.k/kvm/kvm.sh")
      end
    end

    describe "the web process type" do
      let(:web_process) do
        yml.fetch("default_process_types").fetch("web")
      end

      context "when there are no directories containing a project.json" do
        it "should work (the user might be using a custom start command)" do
          expect(out).not_to receive(:fail)
          subject.write_release_yml(buildDir, out)
          expect(File).to exist(yml_path)
        end
      end

      context "when there is a directory with a project.json file" do
        let(:web_dir) do
          File.join(buildDir, "foo").tap { |f| Dir.mkdir(f) }
        end

        let(:project_json) do
          "{}"
        end

        before do
          File.open(File.join(web_dir, "project.json"), 'w') do |f|
            f.write project_json
          end
        end

        it "writes a release yml" do
          subject.write_release_yml(buildDir, out)
          expect(File).to exist(File.join(buildDir, "aspnet5-buildpack-release.yml"))
        end

        it "contains a web process type" do
          expect(yml).to have_key("default_process_types")
          expect(yml.fetch("default_process_types")).to have_key("web")
        end

        it "does not contain any exports (these should be done via .profile.d script)" do
          expect(yml).to have_key("default_process_types")
          expect(yml["default_process_types"]["web"]).not_to include("export")
        end

        context "and the project.json contains a cf-web command" do
          let(:project_json) do
            '{"commands": {"cf-web": "whatever"}}'
          end

          it "changes directory to that directory" do
            expect(web_process).to match("cd foo;")
          end

          it "runs 'k cf-web'" do
            expect(web_process).to match("k cf-web")
          end

          context "and if the cf-web command is empty" do
            let(:project_json) do
              '{"commands": {"cf-web": ""}}'
            end

            it "sets it to serve NoWin.vNext" do
              subject.write_release_yml(buildDir, out)

              json = JSON.parse(IO.read(File.join(web_dir, "project.json")))
              expect(json["commands"]["cf-web"]).to match("Microsoft.AspNet.Hosting --server Nowin.vNext")
            end
          end
        end

        context "and the project.json does not contain a cf-web command" do
          it "adds cf-web command to project.json" do
            subject.write_release_yml(buildDir, out)

            json = JSON.parse(IO.read(File.join(web_dir, "project.json")))
            expect(json).to have_key("commands")
            expect(json["commands"]).to have_key("cf-web")
            expect(json["commands"]["cf-web"]).to match("Microsoft.AspNet.Hosting --server Nowin.vNext")
          end

          context "when Nowin.vNext dependency exists" do
            let(:project_json) do
              '{ "dependencies" : { "Nowin.vNext" : "345" } }'
            end

            it "leaves it alone" do
              subject.write_release_yml(buildDir, out)

              json = JSON.parse(IO.read(File.join(web_dir, "project.json")))
              expect(json["dependencies"]["Nowin.vNext"]).to match("345")
            end
          end

          context "when Nowin.vNext dependency does not exist" do
            it "adds Nowin.vNext dependency to project.json" do
              subject.write_release_yml(buildDir, out)

              json = JSON.parse(IO.read(File.join(web_dir, "project.json")))
              expect(json).to have_key("dependencies")
              expect(json["dependencies"]).to have_key("Nowin.vNext")
              expect(json["dependencies"]["Nowin.vNext"]).to match("1.0.0-*")
            end
          end
        end
      end

      context "when there are multiple directories with a project.json file" do
        let(:web_dir) do
          File.join(buildDir, "foo-cfweb").tap { |f| Dir.mkdir(f) }
        end

        let(:other_dir) do
          File.join(buildDir, "bar").tap { |f| Dir.mkdir(f) }
        end

        context "and one contains a cf-web command" do
          before do
            File.open(File.join(other_dir, "project.json"), 'w') do |f|
              f.write '{ "commands": { "web": "whatever" } }'
            end

            File.open(File.join(web_dir, "project.json"), 'w') do |f|
              f.write '{ "commands": { "cf-web": "whatever" } }'
            end
          end

          it "changes directory to that directory" do
            expect(web_process).to match("cd foo-cfweb;")
          end

          it "runs 'k cf-web'" do
            expect(web_process).to match("k cf-web")
          end
        end

        context "and one contains a web command" do
          before do
            File.open(File.join(other_dir, "project.json"), 'w') do |f|
              f.write '{ "commands": { "something": "whatever" } }'
            end

            File.open(File.join(web_dir, "project.json"), 'w') do |f|
              f.write '{ "commands": { "web": "whatever" } }'
            end
          end

          it "changes directory to that directory" do
            expect(web_process).to match("cd foo-cfweb;")
          end

          it "runs 'k cf-web'" do
            expect(web_process).to match("k cf-web")
          end
        end

        context "and one is Nowin.vNext" do
          let(:nowin_dir) do
            File.join(buildDir, "src", "Nowin.vNext").tap { |f| FileUtils.mkdir_p(f) }
          end

          before do
            File.open(File.join(nowin_dir, "project.json"), 'w') do |f|
              f.write '{ "commands": { "web": "whatever" } }'
            end

            File.open(File.join(web_dir, "project.json"), 'w') do |f|
              f.write '{ "commands": { "whatever": "whatever" } }'
            end
          end

          it "changes directory to the other directory" do
            expect(web_process).to match("cd foo-cfweb;")
          end

          it "runs 'k cf-web'" do
            expect(web_process).to match("k cf-web")
          end
        end
      end
    end
  end
end
