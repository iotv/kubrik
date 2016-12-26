package main

import (
	"fmt"
	"os"
	"github.com/mg4tv/kubrik/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
