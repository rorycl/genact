package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func old() {
	cmd := exec.Command("/home/rory/go/bin/glow", "-")
	r, err := os.ReadFile("README.md")
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stdin = strings.NewReader(string(r))
	err = cmd.Run()
	if err != nil {
		fmt.Println("command error: ", err)
	}
}

func old2() {
	cmd := exec.Command("/home/rory/go/bin/glow", "-w", "100", "-p")
	r, err := os.ReadFile("README.md")
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stdin = strings.NewReader(string(r))
	lsOut, err := cmd.Output()
	if err != nil {
		fmt.Println("output error: ", err)
	}
	fmt.Println(string(lsOut))
	// err = cmd.Run()
	// if err != nil {
	// 	fmt.Println("command error: ", err)
	// }
}

func latest() {
	cmd := exec.Command("/home/rory/go/bin/glow", "-w", "100", "-p")
	r, err := os.ReadFile("README.md")
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stdin = strings.NewReader(string(r))
	cmd.Stdout = os.Stdout
	/*
		lsOut, err := cmd.Output()
		if err != nil {
			fmt.Println("output error: ", err)
		}
		fmt.Println(string(lsOut))
	*/
	err = cmd.Run()
	if err != nil {
		fmt.Println("command error: ", err)
	}

}
