$LOAD_PATH << 'cf_spec'
require 'spec_helper'
require 'rspec'
require 'tmpdir'
require 'fileutils'

describe AspNetCoreBuildpack::DotnetCli do
  let(:installers)        { [ double(:installer, path: '') ] }
  let(:shell)             { AspNetCoreBuildpack.shell }
  let(:out)               { double(:out) }
  let(:build_dir)         { Dir.mktmpdir }
  let(:project_paths)     { [] }
  let(:main_project_path) { 'override' }
  let(:app_dir)           { double(:app_dir, main_project_path: main_project_path,
                                             project_paths: project_paths) }

  subject { described_class.new(build_dir, installers) }

  before do
    allow(AspNetCoreBuildpack).to receive(:shell).and_return(shell)
    allow(AspNetCoreBuildpack::AppDir).to receive(:new).with(build_dir).and_return(app_dir)
    allow(shell).to receive(:exec)
  end

  after do
    FileUtils.rm_rf(build_dir)
  end

  describe '#restore' do
    context 'installed sdk uses msbuild' do
      let(:project_paths) { %w(src/project1/project1.csproj src/project2/project2.csproj) }

      before do
        allow(subject).to receive(:msbuild?).with(build_dir).and_return(true)
      end

      it 'sets up the environent and runs dotnet restore once for each project' do
        expect(shell).to receive(:exec) do |*args|
          cmd = args.first
          expect(cmd).to match(/dotnet restore src\/project1\/project1.csproj/)
        end
        expect(shell).to receive(:exec) do |*args|
          cmd = args.first
          expect(cmd).to match(/dotnet restore src\/project2\/project2.csproj/)
        end

        subject.restore(out)
        expect(shell.env['HOME']).to eq build_dir
        expect(shell.env['LD_LIBRARY_PATH']).to eq "$LD_LIBRARY_PATH:#{build_dir}/libunwind/lib"
        path = "$PATH::#{build_dir}/src/project1/node_modules/.bin:#{build_dir}/src/project2/node_modules/.bin"
        expect(shell.env['PATH']).to eq path
      end
    end

    context 'installed sdk uses project.json' do
      let(:project_paths) { %w(src/project1 src/project2) }

      before do
        allow(subject).to receive(:msbuild?).with(build_dir).and_return(false)
      end

      it 'sets up the environment and runs dotnet restore' do
        expect(shell).to receive(:exec) do |*args|
          cmd = args.first
          expect(cmd).to match(/dotnet restore src\/project1 src\/project2/)
        end

        subject.restore(out)
        expect(shell.env['HOME']).to eq build_dir
        expect(shell.env['LD_LIBRARY_PATH']).to eq "$LD_LIBRARY_PATH:#{build_dir}/libunwind/lib"
        path = "$PATH::#{build_dir}/src/project1/node_modules/.bin:#{build_dir}/src/project2/node_modules/.bin"
        expect(shell.env['PATH']).to eq path
      end
    end
  end

  describe '#publish' do
    context 'installed sdk uses msbuild' do
      let(:main_project_path)      { 'src/project1/project1.csproj' }
      let(:project_paths)          { %w(src/project1/project1.csproj) }
      let(:publish_release_config) { 'override' }

      before do
        @old_env = ENV['PUBLISH_RELEASE_CONFIG']
        ENV['PUBLISH_RELEASE_CONFIG'] = publish_release_config

        allow(subject).to receive(:msbuild?).with(build_dir).and_return(true)
      end

      after do
        ENV['PUBLISH_RELEASE_CONFIG'] = @old_env
      end

      context 'PUBLISH_RELEASE_CONFIG is true' do
        let(:publish_release_config) { 'true' }

        it 'sets up the environment, makes a directory to publish the app, and publishes it' do
          publish_dir = File.join(build_dir, '.cloudfoundry', 'dotnet_publish')
          expect(shell).to receive(:exec) do |*args|
            cmd = args.first
            expect(cmd).to match(/dotnet publish src\/project1\/project1.csproj -o #{publish_dir} -c Release/)
          end

          subject.publish(out)

          expect(File.exist? publish_dir).to be_truthy
          expect(shell.env['HOME']).to eq build_dir
          expect(shell.env['LD_LIBRARY_PATH']).to eq "$LD_LIBRARY_PATH:#{build_dir}/libunwind/lib"
          expect(shell.env['PATH']).to eq "$PATH::#{build_dir}/src/project1/node_modules/.bin"
        end
      end

      context 'PUBLISH_RELEASE_CONFIG is not true' do
        let(:publish_release_config) { nil }

        it 'sets up the environment, makes a directory to publish the app, and publishes it' do
          publish_dir = File.join(build_dir, '.cloudfoundry', 'dotnet_publish')
          expect(shell).to receive(:exec) do |*args|
            cmd = args.first
            expect(cmd).to match(/dotnet publish src\/project1\/project1.csproj -o #{publish_dir} -c Debug/)
          end

          subject.publish(out)

          expect(File.exist? publish_dir).to be_truthy
          expect(shell.env['HOME']).to eq build_dir
          expect(shell.env['LD_LIBRARY_PATH']).to eq "$LD_LIBRARY_PATH:#{build_dir}/libunwind/lib"
          expect(shell.env['PATH']).to eq "$PATH::#{build_dir}/src/project1/node_modules/.bin"
        end
      end
    end
  end
end
