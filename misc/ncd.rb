
require 'json/ext'
require 'json/add/core'

class CheckResultSet
  attr_accessor :results

  @results = []

  def to_json(*a)
    {'json_class'   => self.class.name,
      'results'         => results
    }.to_json(*a)
  end

end

class CheckResult
  attr_accessor :hostname, :servicename, :status, :checkpassive, :checkscheduled, :checkoutput, :start_timestamp, :end_timestamp


  def initialize
    @hostname = "ahost"
  end

end

f = CheckResultSet.new
f.results = [CheckResult.new]

puts f
puts f.to_json
