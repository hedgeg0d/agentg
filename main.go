package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hedgeg0d/agentg/internal/auth"
	"github.com/hedgeg0d/agentg/internal/bot"
	"github.com/hedgeg0d/agentg/internal/config"
	"github.com/hedgeg0d/agentg/internal/store"
)

func main() {
	cfgPath := flag.String("config", "config.json", "path to config file")
	hashPw := flag.Bool("hashpw", false, "read a password from stdin and print its bcrypt hash")
	flag.Parse()

	if *hashPw {
		hashPassword()
		return
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	st, err := store.Open(cfg.DataDir)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	az, err := auth.New(cfg.DataDir, cfg.Access)
	if err != nil {
		log.Fatalf("auth: %v", err)
	}
	if az.PlaintextPassword() {
		log.Print("warning: access password is stored in plaintext; run with -hashpw to generate a bcrypt hash")
	}
	b, err := bot.New(cfg, az, st)
	if err != nil {
		log.Fatalf("bot: %v", err)
	}
	b.Run()
}

func hashPassword() {
	fmt.Fprint(os.Stderr, "Password: ")
	r := bufio.NewReader(os.Stdin)
	line, _ := r.ReadString('\n')
	hash, err := auth.HashPassword(strings.TrimRight(line, "\r\n"))
	if err != nil {
		log.Fatalf("hash: %v", err)
	}
	fmt.Println(hash)
}
