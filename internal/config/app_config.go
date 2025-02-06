package config

import (
	"errors"
	"flag"
	"os"
)

type AppConfig struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

const (
	runAddressEnvName                 = "RUN_ADDRESS"
	dadatabaseURIEnvName              = "DATABASE_URI"
	accruaaccrualSystemAddressEnvName = "ACCRUAL_SYSTEM_ADDRESS"
)

var (
	ErrMissingAddress             = errors.New("missing run address")
	ErrMissingDDURI               = errors.New("missing database URI")
	ErrMissingAccrualSystemAddres = errors.New("missing accrual system address")
)

func NewAppConfig() (*AppConfig, error) {
	var (
		runAddress           string
		databaseURI          string
		accrualSystemAddress string
		errs                 []error
	)

	flag.StringVar(&runAddress, "a", runAddress, "sets host and port to run")
	flag.StringVar(&databaseURI, "d", databaseURI, "set database URI to use")
	flag.StringVar(&accrualSystemAddress, "r", accrualSystemAddress, "set accrual system address and port")
	flag.Parse()

	if runAddressEnv := os.Getenv(runAddressEnvName); runAddressEnv != "" {
		runAddress = runAddressEnv
	}

	if databaseURIEnv := os.Getenv(dadatabaseURIEnvName); databaseURIEnv != "" {
		databaseURI = databaseURIEnv
	}

	if accrualSystemAddressEnv := os.Getenv(accruaaccrualSystemAddressEnvName); accrualSystemAddressEnv != "" {
		accrualSystemAddress = accrualSystemAddressEnv
	}

	if runAddress == "" {
		errs = append(errs, ErrMissingAddress)
	}

	if databaseURI == "" {
		errs = append(errs, ErrMissingDDURI)
	}

	if accrualSystemAddress == "" {
		errs = append(errs, ErrMissingAccrualSystemAddres)
	}

	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	return &AppConfig{
		RunAddress:           runAddress,
		DatabaseURI:          databaseURI,
		AccrualSystemAddress: accrualSystemAddress,
	}, nil
}
