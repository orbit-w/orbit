package mongodbdriver

import "time"

// MongoDBConfig holds MongoDB connection configuration
type MongoDBConfig struct {
	URI                    string        `yaml:"uri" json:"uri" toml:"uri"`
	ConnectTimeout         time.Duration `yaml:"connect_timeout" json:"connect_timeout" toml:"connect_timeout"`
	MaxPoolSize            uint64        `yaml:"max_pool_size" json:"max_pool_size" toml:"max_pool_size"`
	MinPoolSize            uint64        `yaml:"min_pool_size" json:"min_pool_size" toml:"min_pool_size"`
	MaxConnIdleTime        time.Duration `yaml:"max_conn_idle_time" json:"max_conn_idle_time" toml:"max_conn_idle_time"`
	MaxConnecting          uint64        `yaml:"max_connecting" json:"max_connecting" toml:"max_connecting"`
	WriteTimeout           time.Duration `yaml:"write_timeout" json:"write_timeout" toml:"write_timeout"`
	ReadTimeout            time.Duration `yaml:"read_timeout" json:"read_timeout" toml:"read_timeout"`
	RetryWrites            bool          `yaml:"retry_writes" json:"retry_writes" toml:"retry_writes"`
	RetryReads             bool          `yaml:"retry_reads" json:"retry_reads" toml:"retry_reads"`
	PingTimeout            time.Duration `yaml:"ping_timeout" json:"ping_timeout" toml:"ping_timeout"`
	DisconnectTimeout      time.Duration `yaml:"disconnect_timeout" json:"disconnect_timeout" toml:"disconnect_timeout"`
	ServerSelectionTimeout time.Duration `yaml:"server_selection_timeout" json:"server_selection_timeout" toml:"server_selection_timeout"`
}

func (c *MongoDBConfig) DefaultConfig() *MongoDBConfig {
	return &MongoDBConfig{
		URI:             "mongodb://localhost:27017",
		ConnectTimeout:  time.Second * 10,
		MaxPoolSize:     100,
		MinPoolSize:     10,
		MaxConnIdleTime: time.Minute * 5,
		MaxConnecting:   2,
	}
}
