module IniParse
  module Lines
    # A base class from which other line types should inherit.
    module Line
      # ==== Parameters
      # opts<Hash>:: Extra options for the line.
      #
      def initialize(opts = {})
        @comment        = opts.fetch(:comment, nil)
        @comment_sep    = opts.fetch(:comment_sep, ';')
        @comment_prefix = opts.fetch(:comment_prefix, ' ')
        @comment_offset = opts.fetch(:comment_offset, 0)
        @indent         = opts.fetch(:indent, '')
      end

      # Returns if this line has an inline comment.
      def has_comment?
        not @comment.nil?
      end

      # Returns this line as a string as it would be represented in an INI
      # document.
      def to_ini
        [*line_contents].map { |ini|
            if has_comment?
              ini += ' ' if ini =~ /\S/ # not blank
              ini  = ini.ljust(@comment_offset)
              ini += comment
            end
            @indent + ini
          }.join "\n"
      end

      # Returns the contents for this line.
      def line_contents
        ''
      end

      # Returns the inline comment for this line. Includes the comment
      # separator at the beginning of the string.
      def comment
        "#{ @comment_sep }#{ @comment_prefix }#{ @comment }"
      end

      # Returns whether this is a line which has no data.
      def blank?
        false
      end

      # Returns the options used to create the line
      def options
        {
          comment: @comment,
          comment_sep: @comment_sep,
          comment_prefix: @comment_prefix,
          comment_offset: @comment_offset,
          indent: @indent
        }
      end
    end

    # Represents a section header in an INI document. Section headers consist
    # of a string of characters wrapped in square brackets.
    #
    #   [section]
    #   key=value
    #   etc
    #   ...
    #
    class Section
      include Line

      @regex = /^\[        # Opening bracket
                 ([^\]]+)  # Section name
                 \]$       # Closing bracket
               /x

      attr_accessor :key
      attr_reader   :lines

      include Enumerable

      # ==== Parameters
      # key<String>:: The section name.
      # opts<Hash>::  Extra options for the line.
      #
      def initialize(key, opts = {})
        super(opts)
        @key   = key.to_s
        @lines = IniParse::OptionCollection.new
      end

      def self.parse(line, opts)
        if m = @regex.match(line)
          [:section, m[1], opts]
        end
      end

      # Returns this line as a string as it would be represented in an INI
      # document. Includes options, comments and blanks.
      def to_ini
        coll = lines.to_a

        if coll.any?
          [*super,coll.to_a.map do |line|
            if line.kind_of?(Array)
              line.map { |dup_line| dup_line.to_ini }.join($/)
            else
              line.to_ini
            end
          end].join($/)
        else
          super
        end
      end

      # Enumerates through each Option in this section.
      #
      # Does not yield blank and comment lines by default; if you want _all_
      # lines to be yielded, pass true.
      #
      # ==== Parameters
      # include_blank<Boolean>:: Include blank/comment lines?
      #
      def each(*args, &blk)
        @lines.each(*args, &blk)
      end

      # Adds a new option to this section, or updates an existing one.
      #
      # Note that +[]=+ has no knowledge of duplicate options and will happily
      # overwrite duplicate options with your new value.
      #
      #   section['an_option']
      #     # => ['duplicate one', 'duplicate two', ...]
      #   section['an_option'] = 'new value'
      #   section['an_option]
      #     # => 'new value'
      #
      # If you do not wish to overwrite duplicates, but wish instead for your
      # new option to be considered a duplicate, use +add_option+ instead.
      #
      def []=(key, value)
        line = @lines[key.to_s]
        opts = {}
        if line.kind_of?(Array)
          opts = line.first.options
        elsif line.respond_to? :options
          opts = line.options
        end
        @lines[key.to_s] = IniParse::Lines::Option.new(key.to_s, value, opts)
      end

      # Returns the value of an option identified by +key+.
      #
      # Returns nil if there is no corresponding option. If the key provided
      # matches a set of duplicate options, an array will be returned containing
      # the value of each option.
      #
      def [](key)
        key = key.to_s

        if @lines.has_key?(key)
          if (match = @lines[key]).kind_of?(Array)
            match.map { |line| line.value }
          else
            match.value
          end
        end
      end

      # Deletes the option identified by +key+.
      #
      # Returns the section.
      #
      def delete(*args)
        @lines.delete(*args)
        self
      end

      # Like [], except instead of returning just the option value, it returns
      # the matching line instance.
      #
      # Will return an array of lines if the key matches a set of duplicates.
      #
      def option(key)
        @lines[key.to_s]
      end

      # Returns true if an option with the given +key+ exists in this section.
      def has_option?(key)
        @lines.has_key?(key.to_s)
      end

      # Merges section +other+ into this one. If the section being merged into
      # this one contains options with the same key, they will be handled as
      # duplicates.
      #
      # ==== Parameters
      # other<IniParse::Section>:: The section to merge into this one.
      #
      def merge!(other)
        other.lines.each(true) do |line|
          if line.kind_of?(Array)
            line.each { |duplicate| @lines << duplicate }
          else
            @lines << line
          end
        end
      end

      #######
      private
      #######

      def line_contents
        '[%s]' % key
      end
    end

    # Stores options which appear at the beginning of a file, without a
    # preceding section.
    class AnonymousSection < Section
      def initialize
        super('__anonymous__')
      end

      def to_ini
        # Remove the leading space which is added by joining the blank line
        # content with the options.
        super.gsub(/\A\n/, '')
      end

      #######
      private
      #######

      def line_contents
        ''
      end
    end

    # Represents probably the most common type of line in an INI document:
    # an option. Consists of a key and value, usually separated with an =.
    #
    #   key = value
    #
    class Option
      include Line

      @regex = /^\s*([^=]+)  # Option
                 =
                 (.*?)$      # Value
               /x

      attr_accessor :key, :value

      # ==== Parameters
      # key<String>::   The option key.
      # value<String>:: The value for this option.
      # opts<Hash>::    Extra options for the line.
      #
      def initialize(key, value, opts = {})
        super(opts)
        @key, @value = key.to_s, value
      end

      def self.parse(line, opts)
        if m = @regex.match(line)
          [:option, m[1].strip, typecast(m[2].strip), opts]
        end
      end

      # Attempts to typecast values.
      def self.typecast(value)
        case value
          when /^\s*$/                                        then nil
          when /^-?(?:\d|[1-9]\d+)$/                          then Integer(value)
          when /^-?(?:\d|[1-9]\d+)(?:\.\d+)?(?:e[+-]?\d+)?$/i then Float(value)
          when /^true$/i                                      then true
          when /^false$/i                                     then false
          else                                                     value
        end
      end

      #######
      private
      #######

      # returns an array to support multiple lines or a single one at once
      # because of options key duplication
      def line_contents
        if value.kind_of?(Array)
          value.map { |v, i| "#{key} = #{v}" }
        else
          "#{key} = #{value}"
        end
      end
    end

    # Represents a blank line. Used so that we can preserve blank lines when
    # writing back to the file.
    class Blank
      include Line

      def blank?
        true
      end

      def self.parse(line, opts)
        if line !~ /\S/ # blank
          if opts[:comment].nil?
            [:blank]
          else
            [:comment, opts[:comment], opts]
          end
        end
      end
    end

    # Represents a comment. Comment lines begin with a semi-colon or hash.
    #
    #   ; this is a comment
    #   # also a comment
    #
    class Comment < Blank
      # Returns if this line has an inline comment.
      #
      # Being a Comment this will always return true, even if the comment
      # is nil. This would be the case if the line starts with a comment
      # seperator, but has no comment text. See spec/fixtures/smb.ini for a
      # real-world example.
      #
      def has_comment?
        true
      end

      # Returns the inline comment for this line. Includes the comment
      # separator at the beginning of the string.
      #
      # In rare cases where a comment seperator appeared in the original file,
      # but without a comment, just the seperator will be returned.
      #
      def comment
        @comment !~ /\S/ ? @comment_sep : super
      end
    end
  end # Lines
end # IniParse
