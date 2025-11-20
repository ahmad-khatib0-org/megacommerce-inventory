package server

import (
	"github.com/ahmad-khatib0-org/megacommerce-inventory/pkg/models"
	com "github.com/ahmad-khatib0-org/megacommerce-proto/gen/go/common/v1"
	"github.com/spf13/viper"
)

func (s *Server) initSharedConfig() {
	config, err := s.commonClient.ConfigGet()
	if err != nil {
		s.errors <- err
	}

	s.configFn = func() *com.Config { return config }
	s.configMux.Lock()
	s.config = config
	s.configMux.Unlock()
}

func LoadServiceConfig(fileName string) (*models.Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigFile(fileName)
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var c models.Config
	if err := viper.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
