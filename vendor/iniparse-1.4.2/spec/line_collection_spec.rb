require 'spec_helper'

# ----------------------------------------------------------------------------
# Shared specs for all Collection types...
# ----------------------------------------------------------------------------

share_examples_for "LineCollection" do
  before(:each) do
    @collection << (@c1 = IniParse::Lines::Comment.new)
    @collection <<  @i1
    @collection <<  @i2
    @collection << (@b1 = IniParse::Lines::Blank.new)
    @collection <<  @i3
    @collection << (@b2 = IniParse::Lines::Blank.new)
  end

  describe '#each' do
    it 'should remove blanks and comments by default' do
      @collection.each { |l| l.should be_kind_of(@i1.class) }
    end

    it 'should not remove blanks and comments if true is given' do
      arr = []

      # map(true)->each(true) not possible with Enumerable
      @collection.each(true) do |line|
        arr << line
      end

      arr.should == [@c1, @i1, @i2, @b1, @i3, @b2]
    end
  end

  describe '#[]' do
    it 'should fetch the correct value' do
      @collection['first'].should  == @i1
      @collection['second'].should == @i2
      @collection['third'].should  == @i3
    end

    it 'should return nil if the given key does not exist' do
      @collection['does not exist'].should be_nil
    end
  end

  describe '#[]=' do
    it 'should successfully add a new key' do
      @collection['fourth'] = @new
      @collection['fourth'].should == @new
    end

    it 'should successfully update an existing key' do
      @collection['second'] = @new
      @collection['second'].should == @new

      # Make sure the old data is gone.
      @collection.detect { |s| s.key == 'second' }.should be_nil
    end

    it 'should typecast given keys to a string' do
      @collection[:a_symbol] = @new
      @collection['a_symbol'].should == @new
    end
  end

  describe '#<<' do
    it 'should set the key correctly if given a new item' do
      @collection.should_not have_key(@new.key)
      @collection << @new
      @collection.should have_key(@new.key)
    end

    it 'should append Blank lines' do
      @collection << IniParse::Lines::Blank.new
      @collection.instance_variable_get(:@lines).last.should \
        be_kind_of(IniParse::Lines::Blank)
    end

    it 'should append Comment lines' do
      @collection << IniParse::Lines::Comment.new
      @collection.instance_variable_get(:@lines).last.should \
        be_kind_of(IniParse::Lines::Comment)
    end

    it 'should return self' do
      (@collection << @new).should == @collection
    end
  end

  describe '#delete' do
    it 'should remove the given value and adjust the indicies' do
      @collection['second'].should_not be_nil
      @collection.delete('second')
      @collection['second'].should be_nil
      @collection['first'].should == @i1
      @collection['third'].should == @i3
    end

    it "should do nothing if the supplied key does not exist" do
      @collection.delete('does not exist')
      @collection['first'].should == @i1
      @collection['third'].should == @i3
    end
  end

  describe '#to_a' do
    it 'should return an array' do
      @collection.to_a.should be_kind_of(Array)
    end

    it 'should include all lines' do
      @collection.to_a.should == [@c1, @i1, @i2, @b1, @i3, @b2]
    end

    it 'should include references to the same line objects as the collection' do
      @collection << @new
      @collection.to_a.last.object_id.should == @new.object_id
    end
  end

  describe '#to_hash' do
    it 'should return a hash' do
      @collection.to_hash.should be_kind_of(Hash)
    end

    it 'should have the correct keys' do
      hash = @collection.to_hash
      hash.keys.length.should == 3
      hash.should have_key('first')
      hash.should have_key('second')
      hash.should have_key('third')
    end

    it 'should have the correct values' do
      hash = @collection.to_hash
      hash['first'].should  == @i1
      hash['second'].should == @i2
      hash['third'].should  == @i3
    end
  end

  describe '#keys' do
    it 'should return an array of strings' do
      @collection.keys.should == ['first', 'second', 'third']
    end
  end
end

# ----------------------------------------------------------------------------
# On with the collection specs...
# ----------------------------------------------------------------------------

describe 'IniParse::OptionCollection' do
  before(:each) do
    @collection = IniParse::OptionCollection.new
    @i1  = IniParse::Lines::Option.new('first',  'v1')
    @i2  = IniParse::Lines::Option.new('second', 'v2')
    @i3  = IniParse::Lines::Option.new('third',  'v3')
    @new = IniParse::Lines::Option.new('fourth', 'v4')
  end

  it_should_behave_like 'LineCollection'

  describe '#<<' do
    it 'should raise a LineNotAllowed exception if a Section is pushed' do
      lambda { @collection << IniParse::Lines::Section.new('s') }.should \
        raise_error(IniParse::LineNotAllowed)
    end

    it 'should add the Option as a duplicate if an option with the same key exists' do
      option_one = IniParse::Lines::Option.new('k', 'value one')
      option_two = IniParse::Lines::Option.new('k', 'value two')

      @collection << option_one
      @collection << option_two

      @collection['k'].should == [option_one, option_two]
    end
  end

  describe '#keys' do
    it 'should handle duplicates' do
      @collection << @i1 << @i2 << @i3
      @collection << IniParse::Lines::Option.new('first', 'v5')
      @collection.keys.should == ['first', 'second', 'third']
    end
  end
end

describe 'IniParse::SectionCollection' do
  before(:each) do
    @collection = IniParse::SectionCollection.new
    @i1  = IniParse::Lines::Section.new('first')
    @i2  = IniParse::Lines::Section.new('second')
    @i3  = IniParse::Lines::Section.new('third')
    @new = IniParse::Lines::Section.new('fourth')
  end

  it_should_behave_like 'LineCollection'

  describe '#<<' do
    it 'should add merge Section with the other, if it is a duplicate' do
      new_section = IniParse::Lines::Section.new('first')
      @collection << @i1
      @i1.should_receive(:merge!).with(new_section).once
      @collection << new_section
    end
  end
end
