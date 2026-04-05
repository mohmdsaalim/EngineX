docker compose up -d -> run docker
docker compose down ----------> data exist here 
docker compose up ------------> run docker 

docker compsoe down -v stop the docker 

i am using Event-driven microservices architecture. in this project 

day 2 -> installed -> go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

 Why ENUM is important (real-world): --> SQL
	•	Data consistency ✅
	•	Faster comparisons than strings ⚡
	•	Safer logic in backend (Go/Kafka consumers)
// CI pipeline working ✅ passed all the test
// finded the isuue of redis
loading............

// ///////////////////. register user via //////////////////////
grpcurl -plaintext \
  -d '{"email":"test@test.com","password":"password123","full_name":"Test User"}' \
  localhost:9091 \
  auth.v1.AuthService/Register


  //////////////////////// Login /////////////////
  grpcurl -plaintext \
  -d '{"email":"test@test.com","password":"password123"}' \
  localhost:9091 \
  auth.v1.AuthService/Login

/////////////////// validate ///////////////
grpcurl -plaintext \
  -d '{"token":" put token here "}' \
  localhost:9091 \
  auth.v1.AuthService/ValidateToken
  