# Terraform Configuration for PrimoPoker GCP Infrastructure

terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 5.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.9"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.1"
    }
  }
  
  # Comment out backend for initial deployment
  # backend "gcs" {
  #   bucket = "primopoker-terraform-state"
  #   prefix = "terraform/state"
  # }
}

# Variables
variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "Primary GCP region"
  type        = string
  default     = "us-central1"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  default     = "prod"
}

variable "db_password" {
  description = "Database root password"
  type        = string
  sensitive   = true
}

# Locals
locals {
  service_name = "primopoker"
  labels = {
    environment = var.environment
    service     = local.service_name
    managed_by  = "terraform"
  }
}

# Data sources
data "google_project" "project" {
  project_id = var.project_id
}

# Enable required APIs
resource "google_project_service" "required_apis" {
  for_each = toset([
    "cloudbuild.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "compute.googleapis.com",
    "container.googleapis.com",
    "containerregistry.googleapis.com",
    "logging.googleapis.com",
    "monitoring.googleapis.com",
    "pubsub.googleapis.com",
    "redis.googleapis.com",
    "run.googleapis.com",
    "secretmanager.googleapis.com",
    "servicenetworking.googleapis.com",
    "sql-component.googleapis.com",
    "sqladmin.googleapis.com",
    "storage-api.googleapis.com",
    "storage-component.googleapis.com",
    "vpcaccess.googleapis.com",  # Added missing VPC Access API
  ])

  project = var.project_id
  service = each.value

  disable_dependent_services = false
  disable_on_destroy        = false

  # Add explicit timeouts for API enablement
  timeouts {
    create = "10m"
    update = "10m"
  }
}

# Add delay after API enablement
resource "time_sleep" "wait_for_apis" {
  depends_on = [google_project_service.required_apis]
  create_duration = "60s"
}

# Service Account for the application
resource "google_service_account" "app_service_account" {
  account_id   = "${local.service_name}-app"
  display_name = "PrimoPoker Application Service Account"
  description  = "Service account for PrimoPoker application"
  project      = var.project_id

  depends_on = [time_sleep.wait_for_apis]
}

# IAM bindings for the service account
resource "google_project_iam_member" "app_service_account_roles" {
  for_each = toset([
    "roles/cloudsql.client",
    "roles/secretmanager.secretAccessor",
    "roles/pubsub.publisher",
    "roles/pubsub.subscriber",
    "roles/redis.editor",
    "roles/logging.logWriter",
    "roles/monitoring.metricWriter",
    "roles/errorreporting.writer",
  ])

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.app_service_account.email}"
}

# VPC Network
resource "google_compute_network" "vpc_network" {
  name                    = "${local.service_name}-vpc"
  auto_create_subnetworks = false
  project                 = var.project_id
}

# Subnet for the application
resource "google_compute_subnetwork" "app_subnet" {
  name          = "${local.service_name}-subnet"
  ip_cidr_range = "10.0.1.0/24"
  region        = var.region
  network       = google_compute_network.vpc_network.id
  project       = var.project_id

  private_ip_google_access = true
}

# VPC Connector for Cloud Run
resource "google_vpc_access_connector" "connector" {
  name           = "${local.service_name}-connector"
  project        = var.project_id
  region         = var.region
  ip_cidr_range  = "10.8.0.0/28"
  network        = google_compute_network.vpc_network.name
  max_throughput = 300
}

# Cloud SQL Instance
resource "google_sql_database_instance" "postgres_instance" {
  name             = "${local.service_name}-db-${var.environment}"
  database_version = "POSTGRES_15"
  region           = var.region
  project          = var.project_id

  deletion_protection = true

  settings {
    tier                        = "db-custom-2-4096"
    availability_type           = "REGIONAL"
    disk_type                   = "PD_SSD"
    disk_size                   = 100
    disk_autoresize             = true
    disk_autoresize_limit       = 500

    backup_configuration {
      enabled                        = true
      start_time                     = "03:00"
      point_in_time_recovery_enabled = true
      backup_retention_settings {
        retained_backups = 30
        retention_unit   = "COUNT"
      }
    }

    ip_configuration {
      ipv4_enabled                                  = false
      private_network                               = google_compute_network.vpc_network.id
      enable_private_path_for_google_cloud_services = true
    }

    database_flags {
      name  = "log_checkpoints"
      value = "on"
    }

    database_flags {
      name  = "log_connections"
      value = "on"
    }

    database_flags {
      name  = "log_disconnections"
      value = "on"
    }

    database_flags {
      name  = "log_lock_waits"
      value = "on"
    }

    database_flags {
      name  = "log_temp_files"
      value = "0"
    }

    database_flags {
      name  = "log_min_duration_statement"
      value = "1000"
    }

    insights_config {
      query_insights_enabled  = true
      record_application_tags = true
      record_client_address   = true
    }
  }

  depends_on = [google_service_networking_connection.private_vpc_connection]
}

# Private service connection for Cloud SQL
resource "google_compute_global_address" "private_ip_address" {
  name          = "${local.service_name}-private-ip"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.vpc_network.id
  project       = var.project_id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.vpc_network.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}

# Database
resource "google_sql_database" "database" {
  name     = "${local.service_name}_${var.environment}"
  instance = google_sql_database_instance.postgres_instance.name
  project  = var.project_id
}

# Database user
resource "google_sql_user" "database_user" {
  name     = "${local.service_name}_user"
  instance = google_sql_database_instance.postgres_instance.name
  password = var.db_password
  project  = var.project_id
}

# Memorystore Redis instance
resource "google_redis_instance" "redis_instance" {
  name           = "${local.service_name}-redis-${var.environment}"
  tier           = "STANDARD_HA"
  memory_size_gb = 1
  region         = var.region
  project        = var.project_id

  authorized_network = google_compute_network.vpc_network.id
  connect_mode       = "PRIVATE_SERVICE_ACCESS"

  redis_configs = {
    maxmemory-policy = "allkeys-lru"
  }

  labels = local.labels
}

# Pub/Sub topic for real-time messaging
resource "google_pubsub_topic" "game_updates" {
  name    = "${local.service_name}-game-updates"
  project = var.project_id

  labels = local.labels

  message_retention_duration = "86400s" # 24 hours
}

# Pub/Sub subscription for each instance type
resource "google_pubsub_subscription" "game_updates_subscription" {
  name    = "${local.service_name}-game-updates-sub"
  topic   = google_pubsub_topic.game_updates.name
  project = var.project_id

  labels = local.labels

  message_retention_duration = "86400s"
  retain_acked_messages      = false
  ack_deadline_seconds       = 20

  expiration_policy {
    ttl = "86400s"
  }

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "300s"
  }

  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.dead_letter.id
    max_delivery_attempts = 5
  }
}

# Dead letter topic for failed messages
resource "google_pubsub_topic" "dead_letter" {
  name    = "${local.service_name}-dead-letter"
  project = var.project_id

  labels = local.labels
}

# Secret Manager secrets
resource "google_secret_manager_secret" "db_password" {
  secret_id = "${local.service_name}-db-password"
  project   = var.project_id

  labels = local.labels

  replication {
    user_managed {
      replicas {
        location = var.region
      }
      replicas {
        location = "us-east1"
      }
    }
  }
}

resource "google_secret_manager_secret_version" "db_password" {
  secret      = google_secret_manager_secret.db_password.id
  secret_data = var.db_password
}

resource "google_secret_manager_secret" "jwt_secret" {
  secret_id = "${local.service_name}-jwt-secret"
  project   = var.project_id

  labels = local.labels

  replication {
    user_managed {
      replicas {
        location = var.region
      }
      replicas {
        location = "us-east1"
      }
    }
  }
}

resource "google_secret_manager_secret_version" "jwt_secret" {
  secret      = google_secret_manager_secret.jwt_secret.id
  secret_data = base64encode(random_password.jwt_secret.result)
}

# Generate JWT secret
resource "random_password" "jwt_secret" {
  length  = 64
  special = true
}

# Generate random suffix for unique bucket names
resource "random_id" "bucket_suffix" {
  byte_length = 4
}

# Cloud Storage bucket for static assets
resource "google_storage_bucket" "static_assets" {
  name     = "${var.project_id}-${local.service_name}-static-${random_id.bucket_suffix.hex}"
  location = var.region
  project  = var.project_id

  labels = local.labels

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type = "Delete"
    }
  }

  cors {
    origin          = ["*"]
    method          = ["GET", "HEAD"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
}

# Cloud Storage bucket for Terraform state
resource "google_storage_bucket" "terraform_state" {
  name     = "${var.project_id}-terraform-state-${random_id.bucket_suffix.hex}"
  location = var.region
  project  = var.project_id

  labels = local.labels

  uniform_bucket_level_access = true
  
  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      age = 90
    }
    action {
      type = "Delete"
    }
  }
}

# Cloud Armor security policy
resource "google_compute_security_policy" "security_policy" {
  name    = "${local.service_name}-security-policy"
  project = var.project_id

  # Rate limiting rule
  rule {
    action   = "throttle"
    priority = 100
    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = ["*"]
      }
    }
    rate_limit_options {
      conform_action = "allow"
      exceed_action  = "deny(429)"
      enforce_on_key = "IP"
      rate_limit_threshold {
        count        = 100
        interval_sec = 60
      }
    }
    description = "Rate limit: 100 requests per minute per IP"
  }

  # DDoS protection rule
  rule {
    action   = "deny(403)"
    priority = 200
    match {
      expr {
        expression = "origin.region_code == 'CN'"
      }
    }
    description = "Block traffic from certain regions"
  }

  # Default rule
  rule {
    action   = "allow"
    priority = 2147483647
    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = ["*"]
      }
    }
    description = "Default allow rule"
  }
}

# Outputs
output "project_id" {
  description = "GCP Project ID"
  value       = var.project_id
}

output "region" {
  description = "Primary region"
  value       = var.region
}

output "service_account_email" {
  description = "Service account email"
  value       = google_service_account.app_service_account.email
}

output "database_connection_name" {
  description = "Cloud SQL connection name"
  value       = google_sql_database_instance.postgres_instance.connection_name
}

output "database_private_ip" {
  description = "Cloud SQL private IP"
  value       = google_sql_database_instance.postgres_instance.private_ip_address
}

output "redis_host" {
  description = "Redis instance host"
  value       = google_redis_instance.redis_instance.host
}

output "redis_port" {
  description = "Redis instance port"
  value       = google_redis_instance.redis_instance.port
}

output "vpc_connector_name" {
  description = "VPC connector name"
  value       = google_vpc_access_connector.connector.name
}

output "pubsub_topic" {
  description = "Pub/Sub topic name"
  value       = google_pubsub_topic.game_updates.name
}

output "security_policy_name" {
  description = "Cloud Armor security policy name"
  value       = google_compute_security_policy.security_policy.name
}
