package kafka
type Config struct {
    Brokers []string `json:"brokers"`
    Topic   string   `json:"topic"` 
}

func NewConfig(brokers []string, topic string) *Config {
    return &Config{
        Brokers: brokers,
        Topic:   topic,
    }
}