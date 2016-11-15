require 'spec_helper'

# Tests use of the Generator when used like so:
#
#   @gen = IniParse::Generator.new
#   @gen.comment('My very own comment')
#   @gen.section('my_section')
#   @gen.option('my_option', 'my value')
#   ...
#
# Or
#
#   IniParse::Generator.gen do |doc|
#     doc.comment('My very own comment')
#     doc.section('my_section')
#     doc.option('my_option', 'my value')
#   end
#

describe 'When generating a document using Generator without section blocks,' do
  before(:each) { @gen = IniParse::Generator.new }

  # --
  # ==========================================================================
  #   SECTION LINES
  # ==========================================================================
  # ++

  describe 'adding a section' do
    it 'should add a Section to the document' do
      @gen.section("a section")
      @gen.document.should have_section("a section")
    end

    it 'should change the Generator context to the section' do
      @gen.section("a section")
      @gen.context.should == @gen.document['a section']
    end

    it 'should pass extra options to the Section instance' do
      @gen.section("a section", :indent => '    ')
      @gen.document["a section"].to_ini.should match(/\A    /)
    end
  end

  # --
  # ==========================================================================
  #   OPTION LINES
  # ==========================================================================
  # ++

  describe 'adding a option' do
    it 'should pass extra options to the Option instance' do
      @gen.section("a section")
      @gen.option("my option", "a value", :indent => '    ')
      @gen.document["a section"].option("my option").to_ini.should match(/^    /)
    end

    describe 'when the context is a Document' do
      it "should add the option to an __anonymous__ section" do
        @gen.option("key", "value")
        @gen.document['__anonymous__']['key'].should eql('value')
      end
    end

    describe 'when the context is a Section' do
      it 'should add the option to the section' do
        @gen.section("a section")
        @gen.option("my option", "a value")
        @gen.document["a section"].should have_option("my option")
        @gen.document["a section"]["my option"].should == "a value"
      end
    end
  end

  # --
  # ==========================================================================
  #   COMMENT LINES
  # ==========================================================================
  # ++

  describe 'adding a comment' do
    it 'should pass extra options to the Option instance' do
      @gen.comment("My comment", :indent => '    ')
      @gen.document.lines.to_a.first.to_ini.should match(/^    /)
    end

    it 'should ignore any extra :comment option' do
      @gen.comment("My comment", :comment => 'Ignored')
      comment_ini = @gen.document.lines.to_a.first.to_ini
      comment_ini.should match(/My comment/)
      comment_ini.should_not match(/Ignored/)
    end

    describe 'when the context is a Document' do
      it 'should add a comment to the document' do
        @gen.comment('My comment')
        comment = @gen.document.lines.to_a.first
        comment.should be_kind_of(IniParse::Lines::Comment)
        comment.to_ini.should match(/; My comment/)
      end
    end

    describe 'when the context is a Section' do
      it 'should add a comment to the section' do
        @gen.section('a section')
        @gen.comment('My comment')
        comment = @gen.document['a section'].lines.to_a.first
        comment.should be_kind_of(IniParse::Lines::Comment)
        comment.to_ini.should match(/My comment/)
      end
    end
  end

  # --
  # ==========================================================================
  #   BLANK LINES
  # ==========================================================================
  # ++

  describe 'adding a blank line' do
    it 'should add a blank line to the document when it is the context' do
      @gen.blank
      comment = @gen.document.lines.to_a.first
      comment.should be_kind_of(IniParse::Lines::Blank)
    end

    it 'should add a blank line to the section when it is the context' do
      @gen.section('a section')
      @gen.blank
      comment = @gen.document['a section'].lines.to_a.first
      comment.should be_kind_of(IniParse::Lines::Blank)
    end
  end

end
