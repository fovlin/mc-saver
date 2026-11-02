package main

import "acovia.net/record"

func help() error {
	record.Print(helpInfo)
	return nil
}

func replHelp() error {
	record.Print(replHelpInfo)
	return nil
}

var helpInfo = `mc-saver [option] <command> <world> [args...]

backup command:

	run <world> [output]
		back up <world> according to <world>/saver.json.
		output defaults to ".", where a dated zip is created.

	gencfg <world>
		write the default rule file to <world>/saver.json.
		an empty rule file is filled with the defaults, one that already
		has rules is left alone and reported as an error.

	repl [world]
		enter interactive repl mode, for [world] if it is given.

	help
		print this help text.

	about
		print the information and copyright about this kit.

config command:

	indices are 0-based, as shown by the ls commands. an index outside the
	list is an error for both the set and the rm commands.

	on the command line, prefix a negative coordinate with @: write @-3.
	: and = are also accepted as literal prefixes; the prefix is stripped.
	in the repl, plain -3 works.

	ls <world>
		list all dimension rules and file rules.

	ls-config <world> <dimension>...
		list both range rules and simple rules of a dimension.

	ls-dms <world>
		list dimension namespace ids.

	add-dms <world> <dimension>...
		add a dimension with the default range rule.

	rm-dms <world> <dimension>...
		delete a dimension.

	set-dms <world> <old_dimension> <new_dimension>
		rename a dimension, keeping its rules.

	ls-range <world> <dimension>...
		list range rules of a dimension.

	add-range <world> <dimension> <from_x> <from_y> <to_x> <to_y>
		add a range rule to a dimension.

	rm-range <world> <dimension> <index>...
		delete the range rules of the given indices.

	set-range <world> <dimension> <index> <from_x> <from_y> <to_x> <to_y>
		replace the range rule at the given index.

	ls-simple <world> <dimension>...
		list simple rules of a dimension.

	add-simple <world> <dimension> <x> <y>
		add a simple rule to a dimension.

	rm-simple <world> <dimension> <index>...
		delete the simple rules of the given indices.

	set-simple <world> <dimension> <index> <x> <y>
		replace the simple rule at the given index.

	ls-file <world>
		list file rules.

	add-file <world> <name>...
		add one or more file rules.

	rm-file <world> <index>...
		delete the file rules of the given indices.

	set-file <world> <index> <name>
		replace the file rule at the given index.

options:

	-l --legacy:
		legacy world mode, for worlds from before 1.21.11.

	-c --color:
		enable color output.

	-t --time:
		prepend a timestamp to every log line.

	--help:
		print this help text.

`

var replHelpInfo = `repl commands, the selected world is implicit:

	run [output]                back up the selected world
	gencfg                      write the default rule file
	ls                          list all dimension rules and file rules
	ls-config <dimension>...    list range and simple rules
	ls-dms                      list dimension namespace ids
	add-dms <dimension>...      add dimensions with the default range rule
	rm-dms <dimension>...       delete dimensions
	set-dms <old> <new>         rename a dimension, keeping its rules
	ls-range <dimension>...     list range rules
	add-range <dimension> <from_x> <from_y> <to_x> <to_y>
	rm-range <dimension> <index>...
	set-range <dimension> <index> <from_x> <from_y> <to_x> <to_y>
	ls-simple <dimension>...    list simple rules
	add-simple <dimension> <x> <y>
	rm-simple <dimension> <index>...
	set-simple <dimension> <index> <x> <y>
	ls-file                     list file rules
	add-file <name>...
	rm-file <index>...
	set-file <index> <name>

	world                       print the selected world
	select <world>              switch to another world
	help                        print this help text
	about                       print the information and copyright
	exit                        leave repl mode

`
var askWorldHelpInfo = `repl selecter commands:

	select <world>              switch to another world
	help                        print this help text
	exit                        leave repl mode

`
