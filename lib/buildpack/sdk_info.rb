module AspNetCoreBuildpack
  module SdkInfo
    def msbuild?
      msbuild_sdk_versions.include? installed_sdk_version
    end

    def project_json?
      project_json_sdk_versions.include? installed_sdk_version
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

    private

    def installed_sdk_version
      version_file = DotnetSdkInstaller.new(@build_dir, 'dotnet', @deps_dir, @deps_idx, '', '').version_file
      return nil unless File.exist? version_file
      File.read(version_file).strip
    end
  end
end
