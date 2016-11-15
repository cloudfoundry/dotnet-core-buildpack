$:.push File.join(File.dirname(__FILE__), '..', 'lib')

require 'rubygems'
require 'rspec'

require 'iniparse'
require File.join(File.dirname(__FILE__), 'spec_fixtures')

module IniParse
  module Test
    module Helpers
      # Taken from Merb Core's spec helper.
      # Merb is licenced using the MIT License and is copyright
      # Engine Yard Inc.
      class BeKindOf
        def initialize(expected) # + args
          @expected = expected
        end

        def matches?(target)
          @target = target
          @target.kind_of?(@expected)
        end

        def failure_message
          "expected #{@expected} but got #{@target.class}"
        end

        def negative_failure_message
          "expected #{@expected} to not be #{@target.class}"
        end

        def description
          "be_kind_of #{@target}"
        end
      end

      # Used to match line tuples returned by Parser.parse_line.
      class BeLineTuple
        def initialize(type, value_keys = [], *expected)
          @expected_type = type
          @value_keys    = value_keys

          if expected.nil?
            @expected_opts   = {}
            @expected_values = []
          else
            @expected_opts   = expected.pop
            @expected_values = expected
          end
        end

        def matches?(tuple)
          @tuple = tuple

          @failure_message = catch(:fail) do
            tuple?
            correct_type?
            correct_values?
            correct_opts?
            true
          end

          @failure_message == true
        end

        def failure_message
          "expected #{@expected_type} tuple #{@failure_message}"
        end

        def negative_failure_message
          "expected #{@tuple.inspect} to not be #{@expected_type} tuple"
        end

        def description
          "be_#{@expected_type}_tuple #{@tuple}"
        end

        #######
        private
        #######

        # Matchers.

        def tuple?
          throw :fail, "but was #{@tuple.class}" unless @tuple.kind_of?(Array)
        end

        def correct_type?
          throw :fail, "but was #{type} tuple" unless type == @expected_type
        end

        def correct_values?
          # Make sure the values match.
          @value_keys.each_with_index do |key, i|
            if @expected_values[i] != :any && values[i] != @expected_values[i]
              throw :fail, 'with %s value of "%s" but was "%s"' % [
                key, values[i], @expected_values[i]
              ]
            end
          end
        end

        def correct_opts?
          if(! @expected_opts.nil?)
            if (! @expected_opts.empty?) && opts.empty?
              throw :fail, 'with options but there were none'
            end

            @expected_opts.each do |key, value|
              unless opts.has_key?(key)
                throw :fail, 'with option "%s", but key was not present' % key
              end

              unless opts[key] == value
                throw :fail, 'with option "%s" => "%s" but was "%s"' % [
                  key, value, opts[key]
                ]
              end
            end
          end
        end

        # Tuple values, etc.

        def type
          @type ||= @tuple.first
        end

        def values
          @values ||= @tuple.length < 3 ? [] : @tuple[1..-2]
        end

        def opts
          @opts ||= @tuple.last
        end
      end

      def be_kind_of(expected) # + args
        BeKindOf.new(expected)
      end

      def be_section_tuple(key = :any, opts = {})
        BeLineTuple.new(:section, [:key], key, opts)
      end

      def be_option_tuple(key = :any, value = :any, opts = {})
        BeLineTuple.new(:option, [:key, :value], key, value, opts)
      end

      def be_blank_tuple
        BeLineTuple.new(:blank)
      end

      def be_comment_tuple(comment = :any, opts = {})
        BeLineTuple.new(:comment, [:comment], comment, opts)
      end

      def fixture(fix)
        IniParse::Test::Fixtures[fix]
      end
    end
  end
end

RSpec.configure do |config|
  config.include(IniParse::Test::Helpers)
end
