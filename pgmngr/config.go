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

	// migration
	if cfg.Connection.Migration.PingIntervals == 0 {
		cfg.Connection.Migration.PingIntervals = 5
	}

	if cfg.Connection.Migration.Host == "" {
		cfg.Connection.Migration.Host = "localhost"
	}

	if cfg.Connection.Migration.Port == 0 {
		cfg.Connection.Migration.Port = 5432
	}

	if cfg.Connection.Migration.QueryParams == nil {
		// check is sslmode is configured
		cfg.Connection.Migration.QueryParams = map[string]string{
			"sslmode": "disable",
		}
	}

	if len(cfg.Connection.Migration.QueryParams) > 0 {
		// check is sslmode is configured
		sslmodeFound := false
		for k := range cfg.Connection.Migration.QueryParams {
			if k == "sslmode" {
				sslmodeFound = true
			}

		}
		if !sslmodeFound {
			cfg.Connection.Migration.QueryParams["sslmode"] = "disable"
		}
	}

	if cfg.Migration.Table.Schema == "" {
		cfg.Migration.Table.Schema = "public"
	}

	if cfg.Migration.Table.Name == "" {
		cfg.Migration.Table.Name = "schema_migrations"
	}

	// admin defaults are lifted from the migration config
	if cfg.Connection.Admin.PingIntervals == 0 {
		cfg.Connection.Admin.PingIntervals = cfg.Connection.Migration.PingIntervals
	}

	if cfg.Connection.Admin.Host == "" {
		cfg.Connection.Admin.Host = cfg.Connection.Migration.Host
	}

	if cfg.Connection.Admin.Port == 0 {
		cfg.Connection.Admin.Port = cfg.Connection.Migration.Port
	}

	if cfg.Connection.Admin.Database == "" {
		cfg.Connection.Admin.Database = "template1"
	}

	if cfg.Connection.Admin.QueryParams == nil {
		// check is sslmode is configured
		cfg.Connection.Admin.QueryParams = map[string]string{
			"sslmode": "disable",
		}
	}

	if len(cfg.Connection.Admin.QueryParams) > 0 {
		// check is sslmode is configured
		sslmodeFound := false
		for k := range cfg.Connection.Admin.QueryParams {
			if k == "sslmode" {
				sslmodeFound = true
			}

		}
		if !sslmodeFound {
			cfg.Connection.Migration.QueryParams["sslmode"] = "disable"
		}
	}

	return nil
}

// Config stores the options used by pgmngr.
type Config struct {
	Connection struct {
		Admin struct {
			Username      string            `json:"username,omitempty"`
			Password      string            `json:"password,omitempty"`
			Database      string            `json:"template_database,omitempty"`
			Host          string            `json:"host,omitempty"`
			Port          int               `json:"port,omitempty"`
			QueryParams   map[string]string `json:"query_params,omitempty"`
			PingIntervals int               `json:"ping_intervals,omitempty"`
		} `json:"admin"`
		Migration struct {
			PingIntervals int               `json:"ping_intervals,omitempty"`
			Username      string            `json:"username,omitempty"`
			Password      string            `json:"password,omitempty"`
			Database      string            `json:"database,omitempty"`
			Host          string            `json:"host,omitempty"`
			Port          int               `json:"port,omitempty"`
			QueryParams   map[string]string `json:"query_params,omitempty"`
		} `json:"migration"`
	} `json:"connection"`
	Migration struct {
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
	u.Host = c.Connection.Migration.Host
	if c.Connection.Migration.Port != 0 {
		u.Host = strings.Join(
			[]string{c.Connection.Migration.Host, strconv.Itoa(c.Connection.Migration.Port)},
			":",
		)
	}
	u.User = url.UserPassword(
		c.Connection.Migration.Username,
		c.Connection.Migration.Password,
	)

	p, err := url.Parse(c.Connection.Migration.Database)
	if err != nil {
		return "", NewError(err)
	}
	u.Path = path.Join(u.Path, p.Path)

	if len(c.Connection.Migration.QueryParams) > 0 {
		qp := url.Values{}
		for k, v := range c.Connection.Migration.QueryParams {
			qp.Add(k, v)
		}
		if len(qp) > 0 {
			u.RawQuery = qp.Encode()
		}
	}

	return u.String(), nil
}

func (c *Config) dbAdminURL() (string, error) {
	u := url.URL{}
	u.Scheme = postgresScheme
	u.Host = c.Connection.Admin.Host
	if c.Connection.Admin.Port != 0 {
		u.Host = strings.Join(
			[]string{c.Connection.Admin.Host, strconv.Itoa(c.Connection.Admin.Port)},
			":",
		)
	}
	u.User = url.UserPassword(
		c.Connection.Admin.Username,
		c.Connection.Admin.Password,
	)

	p, err := url.Parse(c.Connection.Admin.Database)
	if err != nil {
		return "", NewError(err)
	}
	u.Path = path.Join(u.Path, p.Path)

	if len(c.Connection.Admin.QueryParams) > 0 {
		qp := url.Values{}
		for k, v := range c.Connection.Admin.QueryParams {
			qp.Add(k, v)
		}
		if len(qp) > 0 {
			u.RawQuery = qp.Encode()
		}
	}

	return u.String(), nil
}
