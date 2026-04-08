// migration - db models
// db-model sqlc + pgxpool
// global err and rsponce 
// .env 
// kafka topics 3 
// docker redis, kafka, postgres, -> 6 services(not connected)
// auth via grpc
// risk via grpc
// gateway http 
// engine kafka



commands

1. to check user inisde dock-> pg
docker exec -it engine_postgres psql -U engine_user -d engine_db \
  -c "SELECT id, email, full_name, created_at FROM users;"

2. docker exec -it engine_redis redis-cli keys "session:*"
3. 