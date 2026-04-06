package kafka
//  Currently my gateway/kafka_producer.go creates its own kafka.Writer directly with Async: false and RequireAcks: RequireOne, which is a performance bottleneck.
// I want to:

// Refactor so internal/kafka/producer.go is the single reusable producer
// Switch to Async: true with proper error handling via error callback
// Have gateway/kafka_producer.go reuse that base producer instead of creating its own writer
// Maintain partition ordering by symbol key (Hash balancer — keep this)

// Show me the refactored code for both files following industry standard high-performance Kafka producer patterns in Go."