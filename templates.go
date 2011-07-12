package main

import (
	"template"
	"strings"
	"strconv"
	"io"
)

const activeCheck = `### Active Check Result File ###
file_time={TimeNow}

{.section ServiceName}
### Nagios Service Check Result ###
{.or}
### Nagios Host Check Result ###
{.end}
# Time: Wed Jun  1 16:58:28 2011
host_name={Hostname}
{.section ServiceName}
service_description={@}
{.end}
check_type={CheckPassive}
check_options=0
scheduled_check=1
reschedule_check=1
latency=0.210000
start_time={StartTimestamp}
finish_time={EndTimestamp}
early_timeout=0
exited_ok=1
return_code={Status}
output={CheckOutput}
`

var fmap = template.FormatterMap{
	"escstr": escapedStringFormatter,
}

func escapedStringFormatter(w io.Writer, format string, value ...interface{}) {
	template.StringFormatter(w, format, strings.Trim(strconv.Quote(value[0].(string)), "\""))
}


var tActiveCheck = template.MustParse(activeCheck, nil)
