package types

type (
	ModuleConfigConnector struct {
		Type string `json:"type"` // "rest", "db", "graphql", "elasticsearch", "mongodb", "kafka", "redis", "grpc", "1c-odata"

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

		// 1C:Enterprise OData connector (type "1c-odata"). RestURL is the
		// OData service root, e.g. http://host/base/odata/standard.odata;
		// RestHeaders/RestDataPath are still honored (RestDataPath defaults
		// to "value", 1C's standard OData JSON envelope, when empty).
		OneCEntity   string `json:"oneCEntity,omitempty"` // e.g. Catalog_Номенклатура
		OneCSelect   string `json:"oneCSelect,omitempty"` // OData $select
		OneCFilter   string `json:"oneCFilter,omitempty"` // OData $filter
		OneCTop      int    `json:"oneCTop,omitempty"`    // page size, default 100
		OneCUsername string `json:"oneCUsername,omitempty"`
		OneCPassword string `json:"oneCPassword,omitempty"` // plaintext, or via SecretRefs["oneCPassword"]

		FieldMapping ConnectorFieldMappingSet `json:"fieldMapping,omitempty"`

		// SecretRefs points sensitive fields at a Vault secret instead of
		// storing them in plaintext above. Keys are logical field names:
		// "dbConnectionString", "redisPass", or "restHeader:<HeaderName>"
		// (e.g. "restHeader:Authorization"). A ref present here always wins
		// over the plaintext field of the same name - see
		// service.resolveConnectorSecrets.
		SecretRefs map[string]string `json:"secretRefs,omitempty"`
	}

	ConnectorFieldMapping struct {
		Field  string `json:"field"`
		Source string `json:"source"`
	}

	ConnectorFieldMappingSet []ConnectorFieldMapping
)
