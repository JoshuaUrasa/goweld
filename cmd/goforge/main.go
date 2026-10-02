package main

import (
	"fmt"
	"os"
)


func main(){

//check if user enter command

if len(os.Args) == 1 {
	fmt.Println("please enter a command")
} else{
	command := os.Args[1]

	//use case for command
	switch command {
		case "version":
		fmt.Println("0.1.0")


		case "new":
		switch {
			case len(os.Args) == 2:
			fmt.Println("please provide a project name")
			case len(os.Args) > 3:
			fmt.Println("please write only the project name")
			default:
			name := os.Args[2]
			if err := New(name); err != nil {
				fmt.Println("error:", err)
					}
				}
		default:
		fmt.Println("unknown command")
			}


	}


}
