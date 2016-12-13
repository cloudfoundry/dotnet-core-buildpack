module AspNetCoreBuildpack
  module SdkInfo
    def installed_sdk_version(build_dir)
      version_file = DotnetSdkInstaller.new(build_dir, '', '', '').version_file

      raise ".NET SDK version file: #{version_file} does not exist" unless File.exist? version_file
      File.read(version_file).strip
    end

    def msbuild?(build_dir)
      msbuild_sdk_versions.include? installed_sdk_version(build_dir)
    end

    def project_json?(build_dir)
      project_json_sdk_versions.include? installed_sdk_version(build_dir)
    end

    def dotnet_sdk_tools_file
      File.join(File.dirname(__FILE__), '..', '..', 'dotnet-sdk-tools.yml')
    end

    def msbuild_sdk_versions
      YAML.load_file(dotnet_sdk_tools_file)['msbuild']
    end

    def project_json_sdk_versions
      YAML.load_file(dotnet_sdk_tools_file)['project_json']
    end
  end
end
