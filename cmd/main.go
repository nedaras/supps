package main

import (
	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
)

func main() {		
	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer keyboard.Close()

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}
		color.New(color.Bold, color.FgCyan).Printf("You pressed: rune %q, key %X\n", char, key)
    if key == keyboard.KeyCtrlC {
			break
		}
	}	
}
