dir = File.expand_path('iniparse', File.dirname(__FILE__))

require File.join(dir, 'document')
require File.join(dir, 'generator')
require File.join(dir, 'line_collection')
require File.join(dir, 'lines')
require File.join(dir, 'parser')

module IniParse
  VERSION = '1.4.2'

  # A base class for IniParse errors.
  class IniParseError < StandardError; end

  # Raised if an error occurs parsing an INI document.
  class ParseError < IniParseError; end

  # Raised when an option line is found during parsing before the first
  # section.
  class NoSectionError < ParseError; end

  # Raised when a line is added to a collection which isn't allowed (e.g.
  # adding a Section line into an OptionCollection).
  class LineNotAllowed < IniParseError; end

  module_function

  # Parse given given INI document source +source+.
  #
  # See IniParse::Parser.parse
  #
  # ==== Parameters
  # source<String>:: The source from the INI document.
  #
  # ==== Returns
  # IniParse::Document
  #
  def parse(source)
    IniParse::Parser.new(source.gsub(/(?<!\\)\\\n/, '')).parse
  end

  # Opens the file at +path+, reads and parses it's contents.
  #
  # ==== Parameters
  # path<String>:: The path to the INI document.
  #
  # ==== Returns
  # IniParse::Document
  #
  def open(path)
    document = parse(File.read(path))
    document.path = path
    document
  end

  # Creates a new IniParse::Document using the specification you provide.
  #
  # See IniParse::Generator.
  #
  # ==== Returns
  # IniParse::Document
  #
  def gen(&blk)
    IniParse::Generator.new.gen(&blk)
  end
end
