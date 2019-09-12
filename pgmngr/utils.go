package pgmngr

import (
	"database/sql"
	"fmt"
	"time"
)

func pingDatabase(db *sql.DB, dbPingTime int) error {
	for i := 0; i < dbPingTime; i++ {
		err := db.Ping()
		if err == nil {
			return err
		}
		time.Sleep(time.Second)
	}
	return NewError(
		fmt.Errorf("failed to ping the database after %v seconds", dbPingTime),
	)
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
