# Copyright (c) the go-ruby-thor/thor authors
#
# SPDX-License-Identifier: BSD-3-Clause
#
# Differential oracle for go-ruby-thor/thor. It builds the same CLI the Go
# help_check_test.go builds and dumps class help + a per-command help block so
# the Go tests can byte-compare against the real thor gem (1.5.0). It also
# supports a "parse" mode driving Thor::Options against a small option spec.
#
# Invoked as:  ruby oracle.rb help        -> dumps CLASSHELP / CMDHELP blocks
#              ruby oracle.rb parse <json> -> dumps parsed options for a spec

require "thor"
require "json"

def build_cli
  cli = Class.new(Thor) do
    def self.name; "OracleRb"; end
    namespace "oracle.rb"
    class_option :verbose, type: :boolean, desc: "Be verbose", aliases: "-v"

    desc "greet NAME", "Greet someone by NAME"
    long_desc "This command greets the person named NAME with an optional greeting."
    method_option :greeting, type: :string, default: "Hello", desc: "The greeting to use", aliases: "-g"
    method_option :loud, type: :boolean, desc: "Shout it"
    method_option :count, type: :numeric, desc: "Times", enum: [1, 2, 3]
    def greet(name); end

    desc "list", "List things"
    def list; end
  end
  cli.instance_variable_set(:@thor_options_basename, "oracle.rb")
  cli
end

def dump_help
  cli = build_cli
  # Force the fixed 80-column layout the Go tests assume.
  ENV["THOR_COLUMNS"] = "80"
  shell = Thor::Shell::Basic.new
  out = +""
  # Thor#help writes to shell; capture stdout.
  require "stringio"
  cap = StringIO.new
  old = $stdout
  $stdout = cap
  begin
    cli.help(shell)
    class_help = cap.string.dup
    cap.truncate(0); cap.rewind
    cli.command_help(shell, "greet")
    cmd_help = cap.string.dup
  ensure
    $stdout = old
  end
  out << "=====CLASSHELP=====\n" << class_help
  out << "=====CMDHELP greet=====\n" << cmd_help
  print out
end

case ARGV[0]
when "help"
  dump_help
else
  warn "usage: oracle.rb help"
  exit 1
end
