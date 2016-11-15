require 'spec_helper'

# --
# ============================================================================
#   Empty array.
# ============================================================================
# ++

describe 'An empty array' do
  it 'should not pass be_section_tuple' do
    [].should_not be_section_tuple
  end

  it 'should not pass be_option_tuple' do
    [].should_not be_option_tuple
  end

  it 'should not pass be_blank_tuple' do
    [].should_not be_blank_tuple
  end

  it 'should not pass be_comment_tuple' do
    [].should_not be_comment_tuple
  end
end

# --
# ============================================================================
#   Section tuple.
# ============================================================================
# ++

describe 'Line tuple [:section, "key", {:opt => "val"}]' do
  before(:all) { @tuple = [:section, "key", {:opt => "val"}] }

  it 'should pass be_section_tuple' do
    @tuple.should be_section_tuple
  end

  it 'should pass be_section_tuple("key")' do
    @tuple.should be_section_tuple("key")
  end

  it 'should fail be_section_tuple("invalid")' do
    @tuple.should_not be_section_tuple("invalid")
  end

  it 'should pass be_section_tuple("key", {:opt => "val"})' do
    @tuple.should be_section_tuple("key", {:opt => "val"})
  end

  it 'should not pass be_section_tuple("key", {:invalid => "val"})' do
    @tuple.should_not be_section_tuple("key", {:invalid => "val"})
  end

  it 'should not pass be_section_tuple("key", {:opt => "invalid"})' do
    @tuple.should_not be_section_tuple("key", {:opt => "invalid"})
  end

  it 'should fail be_option_tuple' do
    @tuple.should_not be_option_tuple
  end

  it 'should fail be_blank_tuple' do
    @tuple.should_not be_blank_tuple
  end

  it 'should fail be_comment_tuple' do
    @tuple.should_not be_comment_tuple
  end
end

# --
# ============================================================================
#   Option tuple.
# ============================================================================
# ++

describe 'Line tuple [:option, "key", "val", {:opt => "val"}]' do
  before(:all) { @tuple = [:option, "key", "val", {:opt => "val"}] }

  it 'should fail be_section_tuple' do
    @tuple.should_not be_section_tuple
  end

  it 'should pass be_option_tuple' do
    @tuple.should be_option_tuple
  end

  it 'should pass be_option_tuple("key")' do
    @tuple.should be_option_tuple("key")
  end

  it 'should fail be_option_tuple("invalid")' do
    @tuple.should_not be_option_tuple("invalid")
  end

  it 'should pass be_option_tuple("key", "val")' do
    @tuple.should be_option_tuple("key", "val")
  end

  it 'should pass be_option_tuple(:any, "val")' do
    @tuple.should be_option_tuple(:any, "val")
  end

  it 'should fail be_option_tuple("key", "invalid")' do
    @tuple.should_not be_option_tuple("key", "invalid")
  end

  it 'should pass be_option_tuple("key", "val", { :opt => "val" })' do
    @tuple.should be_option_tuple("key", "val", { :opt => "val" })
  end

  it 'should fail be_option_tuple("key", "val", { :invalid => "val" })' do
    @tuple.should_not be_option_tuple("key", "val", { :invalid => "val" })
  end

  it 'should fail be_option_tuple("key", "val", { :opt => "invalid" })' do
    @tuple.should_not be_option_tuple("key", "val", { :opt => "invalid" })
  end

  it 'should fail be_blank_tuple' do
    @tuple.should_not be_blank_tuple
  end

  it 'should fail be_comment_tuple' do
    @tuple.should_not be_comment_tuple
  end
end

# --
# ============================================================================
#   Blank tuple.
# ============================================================================
# ++

describe 'Line tuple [:blank]' do
  before(:all) { @tuple = [:blank] }

  it 'should fail be_section_tuple' do
    @tuple.should_not be_section_tuple
  end

  it 'should fail be_option_tuple' do
    @tuple.should_not be_option_tuple
  end

  it 'should pass be_blank_tuple' do
    @tuple.should be_blank_tuple
  end

  it 'should fail be_comment_tuple' do
    @tuple.should_not be_comment_tuple
  end
end

# --
# ============================================================================
#   Coment tuple.
# ============================================================================
# ++

describe 'Line tuple [:comment, "A comment", {:opt => "val"}]' do
  before(:all) { @tuple = [:comment, "A comment", {:opt => "val"}] }

  it 'should fail be_section_tuple' do
    @tuple.should_not be_section_tuple
  end

  it 'should fail be_option_tuple' do
    @tuple.should_not be_option_tuple
  end

  it 'should fail be_blank_tuple' do
    @tuple.should_not be_blank_tuple
  end

  it 'should pass be_comment_tuple' do
    @tuple.should be_comment_tuple
  end

  it 'should pass be_comment_tuple("A comment")' do
    @tuple.should be_comment_tuple("A comment")
  end

  it 'should fail be_comment_tuple("Invalid")' do
    @tuple.should_not be_comment_tuple("Invalid")
  end

  it 'should pass be_comment_tuple("A comment", {:opt => "val"})' do
    @tuple.should be_comment_tuple("A comment", {:opt => "val"})
  end

  it 'should fail be_comment_tuple("A comment", {:invalid => "val"})' do
    @tuple.should_not be_comment_tuple("A comment", {:invalid => "val"})
  end

  it 'should fail be_comment_tuple("A comment", {:opt => "invalid"})' do
    @tuple.should_not be_comment_tuple("A comment", {:opt => "invalid"})
  end
end
