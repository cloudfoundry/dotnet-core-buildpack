module IniParse
  # Generator provides a means for easily creating new INI documents.
  #
  # Rather than trying to hack together new INI documents by manually creating
  # Document, Section and Option instances, it is preferable to use Generator
  # which will handle it all for you.
  #
  # The Generator is exposed through IniParse.gen.
  #
  #   IniParse.gen do |doc|
  #     doc.section("vehicle") do |vehicle|
  #       vehicle.option("road_side", "left")
  #       vehicle.option("realistic_acceleration", true)
  #       vehicle.option("max_trains", 500)
  #     end
  #
  #     doc.section("construction") do |construction|
  #       construction.option("build_on_slopes", true)
  #       construction.option("autoslope", true)
  #     end
  #   end
  #
  #   # => IniParse::Document
  #
  # This can be simplified further if you don't mind the small overhead
  # which comes with +method_missing+:
  #
  #   IniParse.gen do |doc|
  #     doc.vehicle do |vehicle|
  #       vehicle.road_side = "left"
  #       vehicle.realistic_acceleration = true
  #       vehicle.max_trains = 500
  #     end
  #
  #     doc.construction do |construction|
  #       construction.build_on_slopes = true
  #       construction.autoslope = true
  #     end
  #   end
  #
  #   # => IniParse::Document
  #
  # If you want to add slightly more complicated formatting to your document,
  # each line type (except blanks) takes a number of optional parameters:
  #
  # :comment::
  #   Adds an inline comment at the end of the line.
  # :comment_offset::
  #   Indent the comment. Measured in characters from _beginning_ of the line.
  #   See String#ljust.
  # :indent::
  #   Adds the supplied text to the beginning of the line.
  #
  # If you supply +:indent+, +:comment_sep+, or +:comment_offset+ options when
  # adding a section, the same options will be inherited by all of the options
  # which belong to it.
  #
  #   IniParse.gen do |doc|
  #     doc.section("vehicle",
  #       :comment => "Options for vehicles", :indent  => "    "
  #     ) do |vehicle|
  #       vehicle.option("road_side", "left")
  #       vehicle.option("realistic_acceleration", true)
  #       vehicle.option("max_trains", 500, :comment => "More = slower")
  #     end
  #   end.to_ini
  #
  #       [vehicle] ; Options for vehicles
  #       road_side = left
  #       realistic_acceleration = true
  #       max_trains = 500 ; More = slower
  #
  class Generator
    attr_reader :context
    attr_reader :document

    def initialize(opts = {}) # :nodoc:
      @document   = IniParse::Document.new
      @context    = @document

      @in_section = false
      @opt_stack  = [opts]
    end

    def gen # :nodoc:
      yield self
      @document
    end

    # Creates a new IniParse::Document with the given sections and options.
    #
    # ==== Returns
    # IniParse::Document
    #
    def self.gen(opts = {}, &blk)
      new(opts).gen(&blk)
    end

    # Creates a new section with the given name and adds it to the document.
    #
    # You can optionally supply a block (as detailed in the documentation for
    # Generator#gen) in order to add options to the section.
    #
    # ==== Parameters
    # name<String>:: A name for the given section.
    #
    def section(name, opts = {})
      if @in_section
        # Nesting sections is bad, mmmkay?
        raise LineNotAllowed, "You can't nest sections in INI files."
      end

      # Add to a section if it already exists
      if @document.has_section?(name.to_s())
        @context = @document[name.to_s()]
      else
        @context = Lines::Section.new(name, line_options(opts))
        @document.lines << @context
      end

      if block_given?
        begin
          @in_section = true
          with_options(opts) { yield self }
          @context = @document
          blank()
        ensure
          @in_section = false
        end
      end
    end

    # Adds a new option to the current section.
    #
    # Can only be called as part of a section block, or after at least one
    # section has been added to the document.
    #
    # ==== Parameters
    # key<String>:: The key (name) for this option.
    # value::       The option's value.
    # opts<Hash>::  Extra options for the line (formatting, etc).
    #
    # ==== Raises
    # IniParse::NoSectionError::
    #   If no section has been added to the document yet.
    #
    def option(key, value, opts = {})
      @context.lines << Lines::Option.new(
        key, value, line_options(opts)
      )
    rescue LineNotAllowed
      # Tried to add an Option to a Document.
      raise NoSectionError,
        'Your INI document contains an option before the first section is ' \
        'declared which is not allowed.'
    end

    # Adds a new comment line to the document.
    #
    # ==== Parameters
    # comment<String>:: The text for the comment line.
    #
    def comment(comment, opts = {})
      @context.lines << Lines::Comment.new(
        line_options(opts.merge(:comment => comment))
      )
    end

    # Adds a new blank line to the document.
    def blank
      @context.lines << Lines::Blank.new
    end

    # Wraps lines, setting default options for each.
    def with_options(opts = {}) # :nodoc:
      opts = opts.dup
      opts.delete(:comment)
      @opt_stack.push( @opt_stack.last.merge(opts))
      yield self
      @opt_stack.pop
    end

    def method_missing(name, *args, &blk) # :nodoc:
      if m = name.to_s.match(/(.*)=$/)
        option(m[1], *args)
      else
        section(name.to_s, *args, &blk)
      end
    end

    #######
    private
    #######

    # Returns options for a line.
    #
    # If the context is a section, we use the section options as a base,
    # rather than the global defaults.
    #
    def line_options(given_opts) # :nodoc:
      @opt_stack.last.empty? ? given_opts : @opt_stack.last.merge(given_opts)
    end
  end
end
