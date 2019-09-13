package pgmngr

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/ParaServices/errgo"
)

func pingDatabase(db *sql.DB, cfg Config) error {
	var err error
	for i := 0; i < cfg.Connection.Migration.PingIntervals; i++ {
		err = db.Ping()
		if err == nil {
			return err
		}
		time.Sleep(time.Second)
	}
	errx := errgo.New(err)
	errx.Message = fmt.Sprintf(
		"failed to ping the database: %s on host: %s:%v after %v seconds",
		cfg.Connection.Migration.Database,
		cfg.Connection.Migration.Host,
		cfg.Connection.Migration.Port,
		cfg.Connection.Migration.PingIntervals,
	)
	errx.Details.Add("database", cfg.Connection.Migration.Database)
	errx.Details.Add("host", cfg.Connection.Migration.Host)
	errx.Details.Add("port", strconv.Itoa(cfg.Connection.Migration.Port))
	errx.Details.Add("ping_time", strconv.Itoa(cfg.Connection.Migration.PingIntervals))
	return errx
}

func pingAdminDatabase(db *sql.DB, cfg Config) error {
	var err error
	for i := 0; i < cfg.Connection.Admin.PingIntervals; i++ {
		err = db.Ping()
		if err == nil {
			return err
		}
		time.Sleep(time.Second)
	}
	errx := errgo.New(err)
	errx.Message = fmt.Sprintf(
		"failed to ping the database: %s on host: %s:%v after %v seconds",
		cfg.Connection.Admin.Database,
		cfg.Connection.Admin.Host,
		cfg.Connection.Admin.Port,
		cfg.Connection.Admin.PingIntervals,
	)
	errx.Details.Add("database", cfg.Connection.Admin.Database)
	errx.Details.Add("host", cfg.Connection.Admin.Host)
	errx.Details.Add("port", strconv.Itoa(cfg.Connection.Admin.Port))
	errx.Details.Add("ping_time", strconv.Itoa(cfg.Connection.Admin.PingIntervals))
	return errx
}

// SliceExclusionInts returns
func sliceExclusionInt64s(sliceA, sliceB []int64) ([]int64, []int64) {
	m := make(map[int64]uint8)
	for _, k := range sliceA {
		m[k] |= (1 << 0)
	}
	for _, k := range sliceB {
		m[k] |= (1 << 1)
	}

	var inAButNotB, inBButNotA []int64
	for k, v := range m {
		a := v&(1<<0) != 0
		b := v&(1<<1) != 0
		switch {
		case a && !b:
			inAButNotB = append(inAButNotB, k)
		case !a && b:
			inBButNotA = append(inBButNotA, k)
		}
	}
	return inAButNotB, inBButNotA
}
