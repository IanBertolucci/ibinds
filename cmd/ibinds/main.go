package main

import (
	"flag"
	"fmt"
	"os"

	"ibinds/internal/config"
	"ibinds/internal/profiles"
)

func main() {
	dir := flag.String("dir", "", "override the iRacing profiles directory")
	listProfiles := flag.Bool("list-profiles", false, "list detected profiles and exit")
	flag.Parse()

	config.FlagDir = *dir

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if *listProfiles {
		ps, err := profiles.Scan(cfg.ProfilesDir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		for _, p := range ps {
			fmt.Printf("%s\t%s\n", p.ID, p.Dir)
		}
		return
	}

	fmt.Println("ok")
}
