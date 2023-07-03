package main

import (
	"aim-oscar/config"
	"aim-oscar/db"
	"aim-oscar/models"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
)

func usage() {
	flag.Usage()
	fmt.Printf("commands:\n\tadd <screen_name> <password> <email>\n\tverify <screen_name>\n")
}

func main() {
	configPath := flag.String("config", "", "Path to app config")
	flag.Parse()

	if configPath == nil || *configPath == "" {
		usage()
		os.Exit(1)
	}

	conf, err := config.FromFile(*configPath)
	if err != nil {
		log.Fatalf("could not parse config: %s", err)
	}

	db, err := db.Connect(&conf.DBConfig)
	if err != nil {
		log.Fatalf("could not connect to DB: %s", err)
	}

	ctx := context.Background()
	cmd := flag.Arg(0)

	if cmd == "add" {
		if len(flag.Args()) < 4 {
			log.Println("missing arguments")
			usage()
			os.Exit(1)
		}

		screenName := flag.Arg(1)
		password := flag.Arg(2)
		email := flag.Arg(3)
		user, err := models.CreateUser(ctx, db, screenName, password, email)
		if err != nil {
			log.Fatalf("could not add user: %s", err)
		}

		log.Printf("Added user")

		user.Verified = true
		if err = user.Update(ctx, db, "verified"); err != nil {
			log.Fatalf("could not verify user: %s", err)
		}

		log.Printf("Verified user")
	} else if cmd == "verify" {
		if len(flag.Args()) < 2 {
			log.Println("missing arguments")
			usage()
			os.Exit(1)
		}

		screenName := flag.Arg(1)
		user, err := models.UserByScreenName(ctx, db, screenName)
		if err != nil {
			log.Fatalf("could not get User by Screen Name: %s", err)
		}

		if user.Verified {
			log.Printf("%s already verified", screenName)
			return
		}

		user.Verified = true
		if err = user.Update(ctx, db, "verified"); err != nil {
			log.Fatalf("could not verify user: %s", err)
		}

		log.Printf("Verified %s", screenName)
	}
}
