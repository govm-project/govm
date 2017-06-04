package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

type Composition struct {
	GoVMs []GoVM `yaml:"go-vms"`
}

func (c *Composition) SetUserData() {
	// Inspect the user data and determine if it is a file or a script.
	for id, govm := range c.GoVMs {
		userData, err := filepath.Abs(govm.UserData)
		if err != nil {
			fmt.Printf("Unable to determine %s user data file location: %v\n", govm, err)
			os.Exit(1)
		}
		// Test if the template file exists
		_, err = os.Stat(userData)
		if err != nil {
			//return fmt.Errorf("file %v does not exist", template)
			// Look for a script verifying the shebang
			var validShebang bool
			validShebangs := []string{
				"#cloud-config",
				"#!/bin/sh",
				"#!/bin/bash",
				"#!/usr/bin/env python",
			}
			_, shebang, err := bufio.ScanLines([]byte(govm.UserData), true)
			if err != nil {
				fmt.Println("Can't scan the user data field")
			}
			for _, sb := range validShebangs {
				if string(shebang) == sb {
					c.GoVMs[id].generateUserData = true
					validShebang = true
				}
			}
			if validShebang == false {
				fmt.Printf("Unable to determine the user data content")
				os.Exit(1)
			}

		}
	}

}

func (c *Composition) SetWorkDir() {
	for id, govm := range c.GoVMs {
		fmt.Printf("vm workdir: %v, and type: %t\n", govm.Workdir, govm.Workdir)
		fmt.Println(wdir)
		if govm.Workdir == "" {
			c.GoVMs[id].Workdir = wdir
		}
		fmt.Printf("vm workdir: %v, and type: %t\n", govm.Workdir, govm.Workdir)
	}
}
