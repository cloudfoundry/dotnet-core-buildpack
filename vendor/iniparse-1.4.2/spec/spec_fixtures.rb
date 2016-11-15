module IniParse
  module Test
    class Fixtures
      @@fixtures = {}

      def self.[](fix)
        if @@fixtures.has_key?(fix)
          @@fixtures[fix]
        else
          @@fixtures[fix] = File.read(
            File.join(File.expand_path('fixtures', File.dirname(__FILE__)), fix)
          )
        end
      end

      def self.[]=(fix, val)
        @@fixtures[fix] = val
      end
    end
  end
end

IniParse::Test::Fixtures[:comment_before_section] = <<-FIX.gsub(/^  /, '')
  ; This is a comment
  [first_section]
  key = value
FIX

IniParse::Test::Fixtures[:blank_before_section] = <<-FIX.gsub(/^  /, '')

  [first_section]
  key = value
FIX

IniParse::Test::Fixtures[:blank_in_section] = <<-FIX.gsub(/^  /, '')
  [first_section]

  key = value
FIX

IniParse::Test::Fixtures[:option_before_section] = <<-FIX.gsub(/^  /, '')
  foo = bar
  [foo]
  another = thing
FIX

IniParse::Test::Fixtures[:invalid_line] = <<-FIX.gsub(/^  /, '')
  this line is not valid
FIX

IniParse::Test::Fixtures[:section_with_equals] = <<-FIX.gsub(/^  /, '')
  [first_section = name]
  key = value
  [another_section = a name]
  another = thing
FIX

IniParse::Test::Fixtures[:comment_line] = <<-FIX.gsub(/^  /, '')
  [first_section]
  ; block comment ;
  ; with more lines ;
  key = value
FIX

IniParse::Test::Fixtures[:duplicate_section] = <<-FIX.gsub(/^  /, '')
  [first_section]
  key = value
  another = thing

  [second_section]
  okay = yes

  [first_section]
  third = fourth
  another = again
FIX
