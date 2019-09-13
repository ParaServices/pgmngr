package pgmngr

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/icrowley/fake"
)

var configFile string

func testConfigFile() *Config {
	cfg := Config{}
	cfg.Connection.Migration.Username = testUsername()
	cfg.Connection.Migration.Password = testPassword()
	cfg.Connection.Migration.Host = testDBHost()
	cfg.Connection.Migration.Database = "pgmngr_test_" + fake.Word() + "_" + fake.Word()
	cfg.Connection.Admin.Username = testUsername()
	cfg.Connection.Admin.Password = testPassword()
	cfg.Connection.Admin.Host = testDBHost()
	cfg.Connection.Admin.Database = "postgres"

	return &cfg
}

func TestMain(m *testing.M) {
	var errOK bool
	tmpDir1, err := ioutil.TempDir("", "")
	errOK = mainErrorHandler(err)
	if errOK {
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir1)

	cfg := testConfigFile()
	cfg.Migration.Directory = tmpDir1

	b, err := json.Marshal(cfg)
	errOK = mainErrorHandler(err)
	if errOK {
		os.Exit(1)
	}

	tmpDir2, tmpFile, err := createTempConfigFile(b, "", "", "pgmngr.json")
	if err != nil {
		errOK = mainErrorHandler(err)
		if errOK {
			os.Exit(1)
		}
	}
	defer os.RemoveAll(tmpDir2)
	configFile = tmpFile

	errOK = mainErrorHandler(err)
	if errOK {
		os.Exit(1)
	}

	m.Run()
}

func testDBHost() string {
	v := os.Getenv("PGMNGR_DB_HOST")
	if v == "" {
		return "localhost"
	}
	return v
}

func testUsername() string {
	v := os.Getenv("PGMNGR_USERNAME")
	if v == "" {
		return "pgmngr"
	}
	return v
}

func testPassword() string {
	v := os.Getenv("PGMNGR_PASSWORD")
	if v == "" {
		return "pgmngr"
	}
	return v
}

func mainErrorHandler(err error) (b bool) {
	if err != nil {
		// notice that we're using 1, so it will actually log the where
		// the error happened, 0 = this function, we don't want that.
		pc, fn, line, _ := runtime.Caller(1)

		log.Printf("[error] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), fn, line, err)
		b = true
	}
	return
}

func createTempConfigFile(fileContent []byte, directory, dirPrefix, fileName string) (string, string, error) {
	dir, err := ioutil.TempDir(dirPrefix, directory)
	if err != nil {
		return "", "", err
	}

	tmpfn := filepath.Join(dir, fileName)
	err = ioutil.WriteFile(tmpfn, fileContent, 0666)
	if err != nil {
		return "", "", err
	}
	return dir, tmpfn, nil
}
