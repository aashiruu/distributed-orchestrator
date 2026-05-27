variable "aws_region" {
  type        = string
  description = "Target region for deployment"
  default     = "us-east-1"
}

variable "db_username" {
  type        = string
  description = "Master access handle for PostgreSQL RDS"
  default     = "orchestrator_user"
}

variable "db_password" {
  type        = string
  description = "Master database access protection code string"
  sensitive   = true
}
