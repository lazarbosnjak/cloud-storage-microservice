package config

type cassandra struct {
	Port string
}

func newCassandraConfig() *cassandra {
	return &cassandra{
		Port: getString("CASSANDRA_PORT"),
	}
}
