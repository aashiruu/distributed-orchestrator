terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# Virtual Private Cloud Isolation Layer
resource "aws_vpc" "orchestrator_vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "orchestrator-infrastructure-vpc"
    Environment = "production"
  }
}

# Private Subnets Configuration for Stateful Engines
resource "aws_subnet" "private_subnet_a" {
  vpc_id            = aws_vpc.orchestrator_vpc.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "${var.aws_region}a"
}

resource "aws_subnet" "private_subnet_b" {
  vpc_id            = aws_vpc.orchestrator_vpc.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = "${var.aws_region}b"
}

# DB Subnet group required for hosting Multi-AZ RDS Instances
resource "aws_db_subnet_group" "db_subnets" {
  name       = "orchestrator-db-subnet-group"
  subnet_ids = [aws_subnet.private_subnet_a.id, aws_subnet.private_subnet_b.id]
}

# Secure Stateful Engines Security Group Boundary
resource "aws_security_group" "internal_data_sg" {
  name        = "orchestrator-internal-data-security-group"
  description = "Block untrusted ingress, permit internal service requests"
  vpc_id      = aws_vpc.orchestrator_vpc.id

  ingress {
    description = "PostgreSQL transactional port access"
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"] # Restricted to internal network communication only
  }

  ingress {
    description = "Redis synchronization locking access"
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Managed Stateful Single Source of Truth Engine (PostgreSQL)
resource "aws_db_instance" "postgres_db" {
  identifier             = "orchestrator-production-db"
  allocated_storage      = 20
  max_allocated_storage  = 100 # Auto-scale to survive heavy metrics logging bursts
  engine                 = "postgres"
  engine_version         = "16.1"
  instance_class         = "db.t4g.micro" # Cost-effective modern Graviton computing profile
  db_name                = "job_orchestrator"
  username               = var.db_username
  password               = var.db_password
  db_subnet_group_name   = aws_db_subnet_group.db_subnets.name
  vpc_security_group_ids = [aws_security_group.internal_data_sg.id]
  skip_final_snapshot    = true
}

# Distributed Mutual Exclusion Lock Layer Engine (Redis)
resource "aws_elasticache_cluster" "redis_cache" {
  cluster_id           = "orchestrator-lock-manager"
  engine               = "redis"
  node_type            = "cache.t4g.micro"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis7"
  port                 = 6379
  subnet_group_name    = aws_elasticache_subnet_group.redis_subnets.name
  security_group_ids   = [aws_security_group.internal_data_sg.id]
}

resource "aws_elasticache_subnet_group" "redis_subnets" {
  name       = "orchestrator-redis-subnet-group"
  subnet_ids = [aws_subnet.private_subnet_a.id, aws_subnet.private_subnet_b.id]
}
