package pglistener

// NewConfig создает новую конфигурацию
func NewConfig(host, port, user, password, database string) Config {
	return Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}
}
