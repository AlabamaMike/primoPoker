# terraform/apis.tf
# Standalone configuration for API enablement - run this first if having API issues

# Note: This file is optional and only needed if you're having API enablement issues
# You can delete this file after APIs are enabled, or run it separately

# Uncomment and use this file independently if needed:

# terraform {
#   required_version = ">= 1.0"
#   required_providers {
#     google = {
#       source  = "hashicorp/google"
#       version = "~> 5.0"
#     }
#     time = {
#       source  = "hashicorp/time"
#       version = "~> 0.9"
#     }
#   }
# }

# variable "project_id" {
#   description = "GCP Project ID"
#   type        = string
# }

# # Enable APIs first
# resource "google_project_service" "apis_only" {
#   for_each = toset([
#     "cloudbuild.googleapis.com",
#     "cloudresourcemanager.googleapis.com", 
#     "compute.googleapis.com",
#     "container.googleapis.com",
#     "containerregistry.googleapis.com",
#     "logging.googleapis.com",
#     "monitoring.googleapis.com",
#     "pubsub.googleapis.com",
#     "redis.googleapis.com",
#     "run.googleapis.com",
#     "secretmanager.googleapis.com",
#     "servicenetworking.googleapis.com",
#     "sql-component.googleapis.com",
#     "sqladmin.googleapis.com",
#     "storage-api.googleapis.com",
#     "storage-component.googleapis.com",
#   ])

#   project = var.project_id
#   service = each.value

#   disable_dependent_services = false
#   disable_on_destroy        = false

#   timeouts {
#     create = "10m"
#     update = "10m"
#   }
# }

# # Wait for APIs to be fully enabled
# resource "time_sleep" "wait_for_api_enablement" {
#   depends_on = [google_project_service.apis_only]
#   create_duration = "120s"
# }

# output "apis_enabled" {
#   value = "All APIs have been enabled and are ready"
#   depends_on = [time_sleep.wait_for_api_enablement]
# }
