package config

type cassandra struct {
	Addr string
}

func newCassandraConfig() *cassandra {
	return &cassandra{
		Addr: getString("CASSANDRA_ADDR"),
	}
}
