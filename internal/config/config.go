package config

import (
	"flag"
	"os"
)

type Parameters struct {
	FlagRunAddr string
	DBpath      string
	AccrualPath string
}

func ParseFlags() Parameters {

	var (
		flagRunAddr string
		dbpath      string
		accrualPath string
	)

	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&accrualPath, "r", "", "path accrual system address")
	flag.StringVar(&dbpath, "d", "", "path to database storage")
	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}

	if envaccrual := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envaccrual != "" {
		accrualPath = envaccrual
	}

	if envDBStoragePath := os.Getenv("DATABASE_DSN"); envDBStoragePath != "" {
		dbpath = envDBStoragePath
	}

	return Parameters{
		FlagRunAddr: flagRunAddr,
		DBpath:      dbpath,
		AccrualPath: accrualPath,
	}
}
