package main

import "acovia.net/record"

func about() error {
	record.Println("mc-saver - selective Minecraft world backup tool")
	record.Println("github link: https://github.com/fovlin/mc-saver")
	record.Println("license: GNU General Public License v3")
	return nil
}
