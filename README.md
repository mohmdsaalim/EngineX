docker compose up -d -> run docker

docker compsoe down -v stop the docker 

i am using Event-driven microservices architecture. in this project 

day 2 -> installed -> go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

 Why ENUM is important (real-world): --> SQL
	•	Data consistency ✅
	•	Faster comparisons than strings ⚡
	•	Safer logic in backend (Go/Kafka consumers)


currenlty i am not dockering golang now iam running this in lockally end of the projct i am going to doekrise the project this is the  for that 
///////////////////////// datas for docerksie thegolang////////////////////////////
You are a senior DevOps engineer. I have a Go microservices project called EngineX.

Current running containers:
- engine_postgres (postgres:16-alpine) — port 5432
- engine_redis (redis:7-alpine) — port 6379  
- engine_kafka (apache/kafka:3.7.0) — port 9092 KRaft mode

I need to add all 6 Go services to docker-compose.yml:
- authsvc — ports 8082, 9091
- gateway — port 8080
- engine — port 9093
- executor — port 9094
- wshub — port 8081
- risksvc — port 9092

Dockerfiles are already written at:
- deployments/docker/authsvc.Dockerfile
- deployments/docker/gateway.Dockerfile
- deployments/docker/engine.Dockerfile
- deployments/docker/executor.Dockerfile
- deployments/docker/wshub.Dockerfile
- deployments/docker/risksvc.Dockerfile

Each Dockerfile uses golang:1.25.5-alpine builder + alpine:latest runner.

Requirements:
1. Add all 6 Go services to docker-compose.yml with correct build context and dockerfile path
2. Each service depends_on postgres, redis, kafka with condition: service_healthy
3. Pass environment variables to each service: POSTGRES_DSN, REDIS_ADDR, KAFKA_BROKER, JWT_SECRET, and service-specific ports
4. Use correct network so all containers talk to each other — postgres host must be "engine_postgres" not "localhost" inside containers
5. Each service must have healthcheck on /healthz endpoint
6. Verify build order is correct — infrastructure starts before Go services
7. Show complete final docker-compose.yml
8. Show how to verify all 9 containers are running and healthy
9. Show how to test end-to-end that Go services can reach Postgres, Redis and Kafka inside Docker network

IMPORTANT:
- Go version is 1.25.5
- Running on Apple Silicon arm64
- Project root is the build context
- No services are implemented yet beyond main.go stubs with DB connection
- Do not change Dockerfile contents
- Show exact docker-compose up command and expected output
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////