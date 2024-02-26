package main

import "github.com/fatih/color"

func printReturnError(msg string, err error) error {
	color.Red("%s: %v", msg, err)
	return err
}
