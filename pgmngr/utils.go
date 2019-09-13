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
	for i := 0; i < cfg.Connection.PingIntervals; i++ {
		err = db.Ping()
		if err == nil {
			return err
		}
		time.Sleep(time.Second)
	}
	errx := errgo.New(err)
	errx.Message = fmt.Sprintf(
		"failed to ping the database: %s on host: %s:%v after %v seconds",
		cfg.Connection.Database,
		cfg.Connection.Host,
		cfg.Connection.Port,
		cfg.Connection.PingIntervals,
	)
	errx.Details.Add("database", cfg.Connection.Database)
	errx.Details.Add("host", cfg.Connection.Host)
	errx.Details.Add("port", strconv.Itoa(cfg.Connection.Port))
	errx.Details.Add("ping_time", strconv.Itoa(cfg.Connection.PingIntervals))
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
