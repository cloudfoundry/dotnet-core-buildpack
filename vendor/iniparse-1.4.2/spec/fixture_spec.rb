require 'spec_helper'

describe "IniParse" do
  describe 'openttd.ini fixture' do
    before(:all) do
      @fixture = fixture('openttd.ini')
    end

    it 'should parse without any errors' do
      lambda { IniParse.parse(@fixture) }.should_not raise_error
    end

    it 'should have the correct sections' do
      IniParse.parse(fixture('openttd.ini')).lines.keys.should == [
        'misc', 'music', 'difficulty', 'game_creation', 'vehicle',
        'construction', 'station', 'economy', 'pf', 'order', 'gui', 'ai',
        'locale', 'network', 'currency', 'servers', 'bans', 'news_display',
        'version', 'preset-J', 'newgrf', 'newgrf-static'
      ]
    end

    it 'should have the correct options' do
      # Test the keys from one section.
      doc     = IniParse.parse(@fixture)
      section = doc['misc']

      section.lines.keys.should == [
        'display_opt', 'news_ticker_sound', 'fullscreen', 'language',
        'resolution', 'screenshot_format', 'savegame_format',
        'rightclick_emulate', 'small_font', 'medium_font', 'large_font',
        'small_size', 'medium_size', 'large_size', 'small_aa', 'medium_aa',
        'large_aa', 'sprite_cache_size', 'player_face',
        'transparency_options', 'transparency_locks', 'invisibility_options',
        'keyboard', 'keyboard_caps'
      ]

      # Test some of the options.
      section['display_opt'].should == 'SHOW_TOWN_NAMES|SHOW_STATION_NAMES|SHOW_SIGNS|FULL_ANIMATION|FULL_DETAIL|WAYPOINTS'
      section['news_ticker_sound'].should be_false
      section['language'].should == 'english_US.lng'
      section['resolution'].should == '1680,936'
      section['large_size'].should == 16

      # Test some other options.
      doc['currency']['suffix'].should == '" credits"'
      doc['news_display']['production_nobody'].should == 'summarized'
      doc['version']['version_number'].should == '070039B0'

      doc['preset-J']['gcf/1_other/BlackCC/mauvetoblackw.grf'].should be_nil
      doc['preset-J']['gcf/1_other/OpenGFX/OpenGFX_-_newFaces_v0.1.grf'].should be_nil
    end

    it 'should be identical to the original when calling #to_ini' do
      IniParse.parse(@fixture).to_ini.should == @fixture
    end
  end

  describe 'race07.ini fixture' do
    before(:all) do
      @fixture = fixture('race07.ini')
    end

    it 'should parse without any errors' do
      lambda { IniParse.parse(@fixture) }.should_not raise_error
    end

    it 'should have the correct sections' do
      IniParse.parse(fixture('race07.ini')).lines.keys.should == [
        'Header', 'Race', 'Slot010', 'Slot016', 'Slot013', 'Slot018',
        'Slot002', 'END'
      ]
    end

    it 'should have the correct options' do
      # Test the keys from one section.
      doc     = IniParse.parse(@fixture)
      section = doc['Slot010']

      section.lines.keys.should == [
        'Driver', 'SteamUser', 'SteamId', 'Vehicle', 'Team', 'QualTime',
        'Laps', 'Lap', 'LapDistanceTravelled', 'BestLap', 'RaceTime'
      ]

      # Test some of the options.
      section['Driver'].should == 'Mark Voss'
      section['SteamUser'].should == 'mvoss'
      section['SteamId'].should == 1865369
      section['Vehicle'].should == 'Chevrolet Lacetti 2007'
      section['Team'].should == 'TEMPLATE_TEAM'
      section['QualTime'].should == '1:37.839'
      section['Laps'].should == 13
      section['LapDistanceTravelled'].should == 3857.750244
      section['BestLap'].should == '1:38.031'
      section['RaceTime'].should == '0:21:38.988'

      section['Lap'].should == [
        '(0, -1.000, 1:48.697)',   '(1, 89.397, 1:39.455)',
        '(2, 198.095, 1:38.060)',  '(3, 297.550, 1:38.632)',
        '(4, 395.610, 1:38.031)',  '(5, 494.242, 1:39.562)',
        '(6, 592.273, 1:39.950)',  '(7, 691.835, 1:38.366)',
        '(8, 791.785, 1:39.889)',  '(9, 890.151, 1:39.420)',
        '(10, 990.040, 1:39.401)', '(11, 1089.460, 1:39.506)',
        '(12, 1188.862, 1:40.017)'
      ]

      doc['Header']['Version'].should == '1.1.1.14'
      doc['Header']['TimeString'].should == '2008/09/13 23:26:32'
      doc['Header']['Aids'].should == '0,0,0,0,0,1,1,0,0'

      doc['Race']['AIDB'].should == 'GameData\Locations\Anderstorp_2007\2007_ANDERSTORP.AIW'
      doc['Race']['Race Length'].should == 0.1
    end

    it 'should be identical to the original when calling #to_ini' do
      pending('awaiting presevation (or lack) of whitespace around =') do
        IniParse.parse(@fixture).to_ini.should == @fixture
      end
    end
  end

  describe 'smb.ini fixture' do
    before(:all) do
      @fixture = fixture('smb.ini')
    end

    it 'should parse without any errors' do
      lambda { IniParse.parse(@fixture) }.should_not raise_error
    end

    it 'should have the correct sections' do
      IniParse.parse(@fixture).lines.keys.should == [
        'global', 'printers'
      ]
    end

    it 'should have the correct options' do
      # Test the keys from one section.
      doc     = IniParse.parse(@fixture)
      section = doc['global']

      section.lines.keys.should == [
        'debug pid', 'log level', 'server string', 'printcap name',
        'printing', 'encrypt passwords', 'use spnego', 'passdb backend',
        'idmap domains', 'idmap config default: default',
        'idmap config default: backend', 'idmap alloc backend',
        'idmap negative cache time', 'map to guest', 'guest account',
        'unix charset', 'display charset', 'dos charset', 'vfs objects',
        'os level', 'domain master', 'max xmit', 'use sendfile',
        'stream support', 'ea support', 'darwin_streams:brlm',
        'enable core files', 'usershare max shares', 'usershare path',
        'usershare owner only', 'usershare allow guests',
        'usershare allow full config', 'com.apple:filter shares by access',
        'obey pam restrictions', 'acl check permissions',
        'name resolve order', 'include'
      ]

      section['display charset'].should == 'UTF-8-MAC'
      section['vfs objects'].should == 'darwinacl,darwin_streams'
      section['usershare path'].should == '/var/samba/shares'
    end

    it 'should be identical to the original when calling #to_ini' do
      IniParse.parse(@fixture).to_ini.should == @fixture
    end
  end

  describe 'authconfig.ini fixture' do
    before(:all) do
      @fixture = fixture('authconfig.ini')
    end

    it 'should be identical to the original when calling #to_ini' do
      IniParse.parse(@fixture).to_ini.should == @fixture
    end
  end

  describe 'option before section fixture' do
    before(:all) do
      @fixture = fixture(:option_before_section)
    end

    it 'should be identical to the original when calling #to_ini' do
      IniParse.parse(@fixture).to_ini.should == @fixture
    end
  end
end
