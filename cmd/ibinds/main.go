package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"ibinds/internal/config"
	"ibinds/internal/profiles"
	"ibinds/internal/server"
	"ibinds/internal/simcheck"
)

func main() {
	dir := flag.String("dir", "", "override the iRacing profiles directory")
	listProfiles := flag.Bool("list-profiles", false, "list detected profiles and exit")
	token := flag.String("token", "", "fix the session token (default: random)")
	port := flag.Int("port", 0, "port to listen on (default: 0 = ephemeral)")
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

	srv := server.New(cfg, server.ServerOpts{
		Token:    *token,
		SimCheck: simcheck.Running,
	})

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	url := fmt.Sprintf("http://127.0.0.1:%d/?token=%s", ln.Addr().(*net.TCPAddr).Port, srv.Token())
	fmt.Println(url)

	if err := http.Serve(ln, srv.Handler()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
