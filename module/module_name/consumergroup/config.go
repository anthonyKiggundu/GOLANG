package consumergroup

import (
	"time"
)

type metricsetConfig struct {
	metricsets []string `config:"metricsets"`
  	enabled bool `config:"enabled"`
	period  time.Duration `config:"period"`
	Retries  int                `config:"retries" validate:"min=0"`
	Backoff  time.Duration      `config:"backoff" validate:"min=0"`
//	TLS      *outputs.TLSConfig `config:"ssl"`
//	Username string             `config:"username"`
//	Password string             `config:"password"`
	ClientID string             `config:"client_id"`
	broker_id int `config:"broker_id"`
	Groups []string `config:"groups"`
	Topics []string `config:"topics"`
}

var defaultConfig = metricsetConfig{
	metricsets: []string{},
	enabled: true,
 	period: 1 * time.Second,
 	broker_id: 1002,
	Retries:  3,
	Backoff:  250 * time.Millisecond,
//	TLS:      nil,
//	Username: "",
//	Password: "",
	ClientID: "metricbeat",
	Groups: []string{},
	Topics: []string{},
}

//func (c *metricsetConfig) Validate() error {
//	if c.Username != "" && c.Password == "" {
//		return fmt.Errorf("password must be set when username is configured")
//	}

//	return nil
//}

