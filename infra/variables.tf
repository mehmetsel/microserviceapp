variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name prefix for all resources"
  type        = string
  default     = "microservice"
}

variable "environment" {
  description = "Deployment environment (prod, staging, dev)"
  type        = string
  default     = "prod"
}

variable "instance_type" {
  description = "EC2 instance type for app servers"
  type        = string
  default     = "t3.micro"
}

variable "asg_min_size" {
  description = "Minimum number of app instances"
  type        = number
  default     = 2
}

variable "asg_max_size" {
  description = "Maximum number of app instances"
  type        = number
  default     = 4
}

variable "asg_desired_capacity" {
  description = "Desired number of app instances (set >=2 for HA)"
  type        = number
  default     = 2
}

variable "auth_user" {
  description = "Basic auth username"
  type        = string
  default     = "admin"
}

variable "auth_pass" {
  description = "Basic auth password"
  type        = string
  sensitive   = true
}

variable "cache_ttl_seconds" {
  description = "Redis response cache TTL in seconds"
  type        = number
  default     = 60
}

variable "image_tag" {
  description = "Docker image tag to deploy (e.g. git SHA)"
  type        = string
  default     = "latest"
}
