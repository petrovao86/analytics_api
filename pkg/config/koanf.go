package config

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type koanfReader struct {
	k *koanf.Koanf
}

var _ IReader = (*koanfReader)(nil)

func NewKoanfReader(configPath string) IReader {
	k := koanf.New(".")
	parser, err := getConfigParser(configPath)
	if err != nil {
		log.Fatal(err)
	}
	k.Load(file.Provider(configPath), parser)
	return &koanfReader{k: k}
}

func (cr *koanfReader) Get(field string) (any, bool) {
	v := cr.k.Get(field)
	return v, v != nil
}

func (cr *koanfReader) Sub(field string) (IReader, bool) {
	r := cr.k.Raw()
	s, ok := r[field]
	if !ok {
		return nil, false
	}
	sMap, ok := s.(map[string]any)
	if !ok {
		return nil, false
	}
	k := koanf.New(".")
	k.Load(confmap.Provider(sMap, ""), nil)
	return &koanfReader{k: k}, true
}

func (cr *koanfReader) Map() map[string]any {
	return cr.k.Raw()
}

func getConfigParser(configPath string) (koanf.Parser, error) {
	ext := filepath.Ext(configPath)
	switch ext {
	case ".yaml", ".yml":
		return yaml.Parser(), nil
	case ".json":
		return json.Parser(), nil
	default:
		return nil, fmt.Errorf("unsupported config format %s", ext)
	}
}
