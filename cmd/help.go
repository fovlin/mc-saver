package main

import "fmt"

func help() error {
	fmt.Print(helpInfo)
	return nil
}

func replHelp() error {
	fmt.Print(replHelpInfo)
	return nil
}

var helpInfo = `mc-saver [-l] [-color] <command> <world> [args...]

every command takes the world directory as its first argument, and the rule
file is <world>/saver.json. a missing rule file is created empty.
running without arguments enters repl mode, where the world is selected once
and every following command uses it. a first argument that looks like a path
(starts with "./" or "/") is taken as the world and enters repl mode directly,
as in "mc-saver /srv/minecraft/world".

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

	indices are 0-based, as shown by the list commands. an index outside the
	list is an error for both the mod and the delete commands.

	list <world>
		list all dimension rules and file rules.

	list-config <world> <dimension>...
		list both range rules and simple rules of a dimension.

	list-dms <world>
		list dimension namespace ids.

	add-dms <world> <dimension>...
		add a dimension with the default range rule.

	del-dms <world> <dimension>...
		delete a dimension.

	mod-dms <world> <old_dimension> <new_dimension>
		rename a dimension, keeping its rules.

	list-range <world> <dimension>...
		list range rules of a dimension.

	add-range <world> <dimension> <from_x> <from_y> <to_x> <to_y>
		add a range rule to a dimension.

	del-range <world> <dimension> <index>...
		delete the range rules of the given indices.

	mod-range <world> <dimension> <index> <from_x> <from_y> <to_x> <to_y>
		replace the range rule at the given index.

	list-simple <world> <dimension>...
		list simple rules of a dimension.

	add-simple <world> <dimension> <x> <y>
		add a simple rule to a dimension.

	del-simple <world> <dimension> <index>...
		delete the simple rules of the given indices.

	mod-simple <world> <dimension> <index> <x> <y>
		replace the simple rule at the given index.

	list-file <world>
		list file rules.

	add-file <world> <name>...
		add one or more file rules.

	del-file <world> <index>...
		delete the file rules of the given indices.

	mod-file <world> <index> <name>
		replace the file rule at the given index.

options:

	-l
		legacy world mode, for worlds from before 1.21.11.

	-color
		enable color output.
`

var replHelpInfo = `repl commands, the selected world is implicit:

	run [output]                back up the selected world
	gencfg                      write the default rule file
	list                        list all dimension rules and file rules
	list-config <dimension>...  list range and simple rules
	list-dms                    list dimension namespace ids
	add-dms <dimension>...      add dimensions with the default range rule
	del-dms <dimension>...      delete dimensions
	mod-dms <old> <new>         rename a dimension, keeping its rules
	list-range <dimension>...   list range rules
	add-range <dimension> <from_x> <from_y> <to_x> <to_y>
	del-range <dimension> <index>...
	mod-range <dimension> <index> <from_x> <from_y> <to_x> <to_y>
	list-simple <dimension>...  list simple rules
	add-simple <dimension> <x> <y>
	del-simple <dimension> <index>...
	mod-simple <dimension> <index> <x> <y>
	list-file                   list file rules
	add-file <name>...
	del-file <index>...
	mod-file <index> <name>

	select <world>              switch to another world
	help                        print this help text
	about                       print the information and copyright
	exit                        leave repl mode
`
