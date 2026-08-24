package types

type (
	ModuleConfigConnector struct {
		Type string `json:"type"` // "rest", "db", "graphql", "elasticsearch", "mongodb", "kafka", "redis", "grpc"

		RestURL         string            `json:"restUrl,omitempty"`
		RestMethod      string            `json:"restMethod,omitempty"`
		RestHeaders     map[string]string `json:"restHeaders,omitempty"`
		RestBody        string            `json:"restBody,omitempty"`
		RestDataPath    string            `json:"restDataPath,omitempty"`
		RestPageParam   string            `json:"restPageParam,omitempty"`
		RestLimitParam  string            `json:"restLimitParam,omitempty"`
		RestOffsetParam string            `json:"restOffsetParam,omitempty"`

		DBConnectionID     uint64 `json:"dbConnectionId,string,omitempty"`
		DBDriver           string `json:"dbDriver,omitempty"`
		DBConnectionString string `json:"dbConnectionString,omitempty"`
		DBQuery            string `json:"dbQuery,omitempty"`

		EsIndex string `json:"esIndex,omitempty"`

		MongoHost  string `json:"mongoHost,omitempty"`
		MongoPort  int    `json:"mongoPort,omitempty"`
		MongoDB    string `json:"mongoDb,omitempty"`
		MongoColl  string `json:"mongoCollection,omitempty"`
		MongoQuery string `json:"mongoQuery,omitempty"`

		KafkaBrokers string `json:"kafkaBrokers,omitempty"`
		KafkaTopic   string `json:"kafkaTopic,omitempty"`
		KafkaGroup   string `json:"kafkaGroup,omitempty"`

		RedisHost string `json:"redisHost,omitempty"`
		RedisPort int    `json:"redisPort,omitempty"`
		RedisPass string `json:"redisPass,omitempty"`
		RedisKey  string `json:"redisKey,omitempty"`
		RedisDB   int    `json:"redisDb,omitempty"`

		GrpcAddr    string `json:"grpcAddr,omitempty"`
		GrpcMethod  string `json:"grpcMethod,omitempty"`
		GrpcPayload string `json:"grpcPayload,omitempty"`

		FieldMapping ConnectorFieldMappingSet `json:"fieldMapping,omitempty"`
	}

	ConnectorFieldMapping struct {
		Field  string `json:"field"`
		Source string `json:"source"`
	}

	ConnectorFieldMappingSet []ConnectorFieldMapping
)
