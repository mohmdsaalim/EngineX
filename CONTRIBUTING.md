// migration - db models
// db-model sqlc + pgxpool + pg connection 
// global err and rsponce 
// .env 
// kafka topics 3 
// docker redis, kafka, postgres, -> 6 services(not connected)
// auth via grpc
// risk via grpc
// gateway http 
// engine kafka

currenlty

studing B-TREE DS
core engine visual flow diagram and unserstd proper workflow


commands
  
1. to check user inisde dock-> pg
docker exec -it engine_postgres psql -U engine_user -d engine_db \
  -c "SELECT id, email, full_name, created_at FROM users;"

2. docker exec -it engine_redis redis-cli keys "session:*"
3. 
kafka:
    image: apache/kafka:3.7.0
    container_name: engine_kafka
    restart: unless-stopped
    ports:
      - "9092:9092"
    environment:
      KAFKA_NODE_ID: 1
      KAFKA_PROCESS_ROLES: broker,controller
      KAFKA_CONTROLLER_QUORUM_VOTERS: 1@kafka:9093
      KAFKA_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT
      KAFKA_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_INTER_BROKER_LISTENER_NAME: PLAINTEXT
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"
      KAFKA_NUM_PARTITIONS: 6
      KAFKA_DEFAULT_REPLICATION_FACTOR: 1
      KAFKA_LOG_RETENTION_HOURS: 168
      KAFKA_LOG_DIRS: /var/lib/kafka/data
      CLUSTER_ID: "MkU3OEVBNTcwNTJENDM2Qk"
    volumes:
      - kafka_data:/var/lib/kafka/data
    healthcheck:
      test: ["CMD-SHELL", "/opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list"]
      interval: 15s
      timeout: 10s
      retries: 10
      start_period: 30s
