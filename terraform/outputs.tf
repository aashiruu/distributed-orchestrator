output "database_endpoint" {
  value       = aws_db_instance.postgres_db.endpoint
  description = "Connection string target endpoint for storage cluster connection pools"
}

output "redis_endpoint" {
  value       = aws_elasticache_cluster.redis_cache.cache_nodes[0].address
  description = "Connection string target endpoint for Redis lock management loops"
}
