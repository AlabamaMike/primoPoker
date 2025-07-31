# terraform/minimal.tf
# Minimal configuration for initial deployment

terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "Primary GCP region"
  type        = string
  default     = "us-central1"
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Just create the essential resources first
resource "google_sql_database_instance" "postgres_instance" {
  name             = "primopoker-db-prod"
  database_version = "POSTGRES_15"
  region           = var.region
  project          = var.project_id

  deletion_protection = false  # Set to true in production

  settings {
    tier              = "db-custom-1-3840"  # Smaller instance to start
    availability_type = "ZONAL"             # Single zone to start
    disk_type         = "PD_SSD"
    disk_size         = 20                  # Smaller disk to start

    backup_configuration {
      enabled = true
      start_time = "03:00"
    }

    ip_configuration {
      ipv4_enabled    = true
      authorized_networks {
        value = "0.0.0.0/0"  # Open access initially - restrict later
        name  = "all"
      }
    }
  }
}

resource "google_sql_database" "database" {
  name     = "primopoker_prod"
  instance = google_sql_database_instance.postgres_instance.name
  project  = var.project_id
}

resource "google_sql_user" "database_user" {
  name     = "primopoker_user"
  instance = google_sql_database_instance.postgres_instance.name
  password = var.db_password
  project  = var.project_id
}

output "database_connection_name" {
  description = "Cloud SQL connection name"
  value       = google_sql_database_instance.postgres_instance.connection_name
}

output "database_ip" {
  description = "Database IP address"
  value       = google_sql_database_instance.postgres_instance.ip_address.0.ip_address
}
