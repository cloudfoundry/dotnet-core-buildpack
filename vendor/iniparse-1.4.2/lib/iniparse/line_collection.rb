module IniParse
  # Represents a collection of lines in an INI document.
  #
  # LineCollection acts a bit like an Array/Hash hybrid, allowing arbitrary
  # lines to be added to the collection, but also indexes the keys of Section
  # and Option lines to enable O(1) lookup via LineCollection#[].
  #
  # The lines instances are stored in an array, +@lines+, while the index of
  # each Section/Option is held in a Hash, +@indicies+, keyed with the
  # Section/Option#key value (see LineCollection#[]=).
  #
  module LineCollection
    include Enumerable

    def initialize
      @lines    = []
      @indicies = {}
    end

    # Retrive a value identified by +key+.
    def [](key)
      has_key?(key) ? @lines[ @indicies[key] ] : nil
    end

    # Set a +value+ identified by +key+.
    #
    # If a value with the given key already exists, the value will be replaced
    # with the new one, with the new value taking the position of the old.
    #
    def []=(key, value)
      key = key.to_s

      if has_key?(key)
        @lines[ @indicies[key] ] = value
      else
        @lines << value
        @indicies[key] = @lines.length - 1
      end
    end

    # Appends a line to the collection.
    #
    # Note that if you pass a line with a key already represented in the
    # collection, the old item will be replaced.
    #
    def <<(line)
      line.blank? ? (@lines << line) : (self[line.key] = line) ; self
    end

    alias_method :push, :<<

    # Enumerates through the collection.
    #
    # By default #each does not yield blank and comment lines.
    #
    # ==== Parameters
    # include_blank<Boolean>:: Include blank/comment lines?
    #
    def each(include_blank = false)
      @lines.each do |line|
        if include_blank || ! (line.is_a?(Array) ? line.empty? : line.blank?)
          yield(line)
        end
      end
    end

    # Removes the value identified by +key+.
    def delete(key)
      key = key.key if key.respond_to?(:key)

      unless (idx = @indicies[key]).nil?
        @indicies.delete(key)
        @indicies.each { |k,v| @indicies[k] = v -= 1 if v > idx }
        @lines.delete_at(idx)
      end
    end

    # Returns whether +key+ is in the collection.
    def has_key?(*args)
      @indicies.has_key?(*args)
    end

    # Return an array containing the keys for the lines added to this
    # collection.
    def keys
      map { |line| line.key }
    end

    # Returns this collection as an array. Includes blank and comment lines.
    def to_a
      @lines.dup
    end

    # Returns this collection as a hash. Does not contain blank and comment
    # lines.
    def to_hash
      Hash[ *(map { |line| [line.key, line] }).flatten ]
    end

    alias_method :to_h, :to_hash
  end

  # A implementation of LineCollection used for storing (mostly) Option
  # instances contained within a Section.
  #
  # Since it is assumed that an INI document will only represent a section
  # once, if SectionCollection encounters a Section key already held in the
  # collection, the existing section is merged with the new one (see
  # IniParse::Lines::Section#merge!).
  class SectionCollection
    include LineCollection

    def <<(line)
      if line.kind_of?(IniParse::Lines::Option)
        option = line
        line   = IniParse::Lines::AnonymousSection.new

        line.lines << option if option
      end

      if line.blank? || (! has_key?(line.key))
        super # Adding a new section, comment or blank line.
      else
        self[line.key].merge!(line)
      end

      self
    end
  end

  # A implementation of LineCollection used for storing (mostly) Option
  # instances contained within a Section.
  #
  # Whenever OptionCollection encounters an Option key already held in the
  # collection, it treats it as a duplicate. This means that instead of
  # overwriting the existing value, the value is changed to an array
  # containing the previous _and_ the new Option instances.
  class OptionCollection
    include LineCollection

    # Appends a line to the collection.
    #
    # If you push an Option with a key already represented in the collection,
    # the previous Option will not be overwritten, but treated as a duplicate.
    #
    # ==== Parameters
    # line<IniParse::LineType::Line>:: The line to be added to this section.
    #
    def <<(line)
      if line.kind_of?(IniParse::Lines::Section)
        raise IniParse::LineNotAllowed,
          "You can't add a Section to an OptionCollection."
      end

      if line.blank? || (! has_key?(line.key))
        super # Adding a new option, comment or blank line.
      else
        self[line.key] = [self[line.key], line].flatten
      end

      self
    end

    # Return an array containing the keys for the lines added to this
    # collection.
    def keys
      map { |line| line.kind_of?(Array) ? line.first.key : line.key }
    end
  end
end
