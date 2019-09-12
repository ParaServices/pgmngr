package pgmngr

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
)

type cliContext interface {
	String(s string) string
}

// LoadConfig loads the configuration from a config file, ENV VARs or command
// line arguments
func LoadConfig(ctx cliContext, cfg *Config) error {
	configPath := ctx.String("config-file")

	file, err := os.Open(configPath)
	if err != nil {
		return NewError(err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return NewError(err)
	}

	err = json.Unmarshal(b, cfg)
	if err != nil {
		return NewError(err)
	}

	if cfg.Connection.PingIntervals == 0 {
		cfg.Connection.PingIntervals = 5
	}

	if cfg.Connection.Host == "" {
		cfg.Connection.Host = "localhost"
	}

	if cfg.Connection.Port == 0 {
		cfg.Connection.Port = 5432
	}

	if cfg.Connection.QueryParams == nil {
		// check is sslmode is configured
		cfg.Connection.QueryParams = map[string]string{
			"sslmode": "disable",
		}
	}

	if len(cfg.Connection.QueryParams) > 0 {
		// check is sslmode is configured
		sslmodeFound := false
		for k := range cfg.Connection.QueryParams {
			if k == "sslmode" {
				sslmodeFound = true
			}

		}
		if !sslmodeFound {
			cfg.Connection.QueryParams["sslmode"] = "disable"
		}
	}

	if cfg.Migration.Table.Schema == "" {
		cfg.Migration.Table.Schema = "public"
	}

	if cfg.Migration.Table.Name == "" {
		cfg.Migration.Table.Name = "schema_migrations"
	}

	return nil
}

// Config stores the options used by pgmngr.
type Config struct {
	Connection struct {
		PingIntervals int               `json:"ping_intervals,omitempty"`
		Username      string            `json:"username,omitempty"`
		Password      string            `json:"password,omitempty"`
		Database      string            `json:"database,omitempty"`
		Host          string            `json:"host,omitempty"`
		Port          int               `json:"port,omitempty"`
		QueryParams   map[string]string `json:"query_params,omitempty"`
	} `json:"connection"`
	Migration struct {
		DumpFile  string `json:"dump_file,omitempty"`
		Directory string `json:"directory,omitempty"`
		Table     struct {
			Schema string `json:"schema"`
			Name   string `json:"name"`
		} `json:"table,omitempty"`
	} `json:"migration"`
}

const postgresScheme = "postgres"

func (c *Config) dbURL() (string, error) {
	u := url.URL{}
	u.Scheme = postgresScheme
	u.Host = c.Connection.Host
	if c.Connection.Port != 0 {
		u.Host = strings.Join(
			[]string{c.Connection.Host, strconv.Itoa(c.Connection.Port)},
			":",
		)
	}
	u.User = url.UserPassword(
		c.Connection.Username,
		c.Connection.Password,
	)

	p, err := url.Parse(c.Connection.Database)
	if err != nil {
		return "", NewError(err)
	}
	u.Path = path.Join(u.Path, p.Path)

	if len(c.Connection.QueryParams) > 0 {
		qp := url.Values{}
		for k, v := range c.Connection.QueryParams {
			qp.Add(k, v)
		}
		if len(qp) > 0 {
			u.RawQuery = qp.Encode()
		}
	}

	return u.String(), nil
}
