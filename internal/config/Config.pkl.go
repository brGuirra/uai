// Code generated from Pkl module `uai.Config`. DO NOT EDIT.
package config

import (
	"context"

	"github.com/apple/pkl-go/pkl"
	"github.com/brGuirra/uai/internal/config/environment"
)

type Config struct {
	Host string `pkl:"host"`

	Port uint16 `pkl:"port"`

	Env environment.Environment `pkl:"env"`

	Db *DB `pkl:"db"`

	Cors *Cors `pkl:"cors"`

	Token *Token `pkl:"token"`

	Smtp *SMTP `pkl:"smtp"`

	DefaultUser *User `pkl:"defaultUser"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Config
func LoadFromPath(ctx context.Context, path string) (ret *Config, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := evaluator.Close()
		if err == nil {
			err = cerr
		}
	}()
	ret, err = Load(ctx, evaluator, pkl.FileSource(path))
	return ret, err
}

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Config
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*Config, error) {
	var ret Config
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
