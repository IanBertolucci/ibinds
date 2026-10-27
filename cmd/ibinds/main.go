package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"ibinds/internal/config"
	"ibinds/internal/inputcap"
	"ibinds/internal/profiles"
	"ibinds/internal/server"
	"ibinds/internal/simcheck"
)

const maxLogSize = 5 << 20

func main() {
	dir := flag.String("dir", "", "override the iRacing profiles directory")
	listProfiles := flag.Bool("list-profiles", false, "list detected profiles and exit")
	token := flag.String("token", "", "fix the session token (default: random)")
	port := flag.Int("port", 0, "port to listen on (default: 0 = ephemeral)")
	capture := flag.Bool("capture", false, "capture a button press and exit")
	noBrowser := flag.Bool("no-browser", false, "do not open the browser on startup")
	logPath := flag.String("log", "", "log file path (default: <user config dir>/ibinds/ibinds.log)")
	flag.Parse()

	logger, closeLog, logFile := openLogger(*logPath)
	defer closeLog()

	config.FlagDir = *dir

	cfg, err := config.Load()
	if err != nil {
		logger.Printf("error: %v", err)
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	logger.Printf("iBinds starting: profiles=%s log=%s", cfg.ProfilesDir, logFile)

	if *listProfiles {
		ps, err := profiles.Scan(cfg.ProfilesDir)
		if err != nil {
			logger.Printf("list profiles: %v", err)
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		for _, p := range ps {
			fmt.Printf("%s\t%s\n", p.ID, p.Dir)
		}
		return
	}

	if *capture {
		result, err := inputcap.New().Capture(inputcap.CaptureOpts{})
		if err != nil {
			logger.Printf("capture error: %v", err)
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		if result == nil {
			logger.Printf("capture timeout")
			fmt.Println("timeout: no button pressed")
			os.Exit(1)
		}
		logger.Printf("capture: button=%d device=%s product=%s", result.Button, result.DeviceInstanceGUID, result.ProductGUID)
		fmt.Printf("button=%d device=%s product=%s\n", result.Button, result.DeviceInstanceGUID, result.ProductGUID)
		return
	}

	srv := server.New(cfg, server.ServerOpts{
		Token:         *token,
		SimCheck:      simcheck.Running,
		InputCapturer: inputcap.New(),
		Logger:        logger,
	})

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		logger.Printf("listen: %v", err)
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	url := fmt.Sprintf("http://127.0.0.1:%d/?token=%s", ln.Addr().(*net.TCPAddr).Port, srv.Token())
	logger.Printf("listening on 127.0.0.1:%d", ln.Addr().(*net.TCPAddr).Port)
	fmt.Println(url)

	if !*noBrowser {
		if err := openBrowser(url); err != nil {
			logger.Printf("open browser: %v", err)
			fmt.Fprintln(os.Stderr, "warning: could not open browser:", err)
		}
	}

	if err := http.Serve(ln, srv.Handler()); err != nil {
		logger.Printf("serve: %v", err)
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// openLogger writes to stderr and a bounded log file. If the file cannot be
// opened, it falls back to stderr only.
func openLogger(path string) (*log.Logger, func(), string) {
	if path == "" {
		if p, err := config.LogPath(); err == nil {
			path = p
		}
	}
	if path == "" {
		return log.New(os.Stderr, "", log.LstdFlags), func() {}, "(stderr only)"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not create log dir: %v\n", err)
		return log.New(os.Stderr, "", log.LstdFlags), func() {}, "(stderr only)"
	}
	if info, err := os.Stat(path); err == nil && info.Size() > maxLogSize {
		_ = os.Rename(path, path+".old")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not open log file: %v\n", err)
		return log.New(os.Stderr, "", log.LstdFlags), func() {}, "(stderr only)"
	}
	return log.New(io.MultiWriter(os.Stderr, f), "", log.LstdFlags), func() { _ = f.Close() }, path
}
