#!/bin/bash

# 🚀 Enhanced PrimoPoker GCP Deployment Script with API Handling
# This script handles API enablement issues gracefully

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID="${PROJECT_ID:-}"
REGION="${REGION:-us-central1}"
SERVICE_NAME="primopoker"
TERRAFORM_DIR="./terraform"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

enable_apis_manually() {
    log_info "Enabling GCP APIs manually..."
    
    local apis=(
        "cloudbuild.googleapis.com"
        "compute.googleapis.com"
        "secretmanager.googleapis.com"
        "sqladmin.googleapis.com"
        "run.googleapis.com"
        "containerregistry.googleapis.com"
        "redis.googleapis.com"
        "pubsub.googleapis.com"
        "logging.googleapis.com"
        "monitoring.googleapis.com"
        "servicenetworking.googleapis.com"
        "storage-api.googleapis.com"
    )
    
    for api in "${apis[@]}"; do
        log_info "Enabling $api..."
        if gcloud services enable "$api" --project="$PROJECT_ID"; then
            log_success "✓ $api enabled"
        else
            log_warning "Failed to enable $api - may already be enabled"
        fi
    done
    
    log_info "Waiting 60 seconds for API propagation..."
    sleep 60
    
    log_success "APIs enabled and ready"
}

verify_apis() {
    log_info "Verifying API enablement..."
    
    local required_apis=("compute" "secretmanager" "sqladmin")
    
    for api in "${required_apis[@]}"; do
        if gcloud services list --enabled --filter="name:$api" --format="value(name)" | grep -q "$api"; then
            log_success "✓ $api API is enabled"
        else
            log_error "✗ $api API is not enabled"
        fi
    done
}

deploy_infrastructure_staged() {
    log_info "Deploying infrastructure in stages..."
    
    cd "$TERRAFORM_DIR"
    
    # Initialize Terraform
    terraform init
    
    # Stage 1: Enable APIs via Terraform (with retry)
    log_info "Stage 1: Enabling APIs via Terraform..."
    local max_retries=3
    local retry_count=0
    
    while [ $retry_count -lt $max_retries ]; do
        if terraform apply -target=google_project_service.required_apis -auto-approve -var="project_id=$PROJECT_ID" -var="region=$REGION"; then
            log_success "APIs enabled via Terraform"
            break
        else
            retry_count=$((retry_count + 1))
            log_warning "Terraform API enablement failed, attempt $retry_count/$max_retries"
            if [ $retry_count -lt $max_retries ]; then
                log_info "Waiting 30 seconds before retry..."
                sleep 30
            fi
        fi
    done
    
    if [ $retry_count -eq $max_retries ]; then
        log_warning "Terraform API enablement failed, using manual method..."
        enable_apis_manually
    fi
    
    # Wait for API propagation
    log_info "Waiting 60 seconds for API propagation..."
    sleep 60
    
    # Stage 2: Deploy remaining infrastructure
    log_info "Stage 2: Deploying remaining infrastructure..."
    terraform apply -auto-approve -var="project_id=$PROJECT_ID" -var="region=$REGION"
    
    cd ..
    log_success "Infrastructure deployment completed"
}

# Main execution
main() {
    echo
    log_info "🚀 Starting Enhanced PrimoPoker GCP Deployment"
    echo
    
    # Check prerequisites
    if [ -z "$PROJECT_ID" ]; then
        log_error "PROJECT_ID environment variable is not set"
    fi
    
    # Verify gcloud authentication
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
        log_error "Not authenticated with gcloud. Run 'gcloud auth login'"
    fi
    
    # Set the project
    gcloud config set project "$PROJECT_ID"
    
    # Enable APIs manually first as backup
    enable_apis_manually
    
    # Verify APIs are working
    verify_apis
    
    # Deploy infrastructure with staged approach
    deploy_infrastructure_staged
    
    echo
    log_success "🎉 Deployment completed successfully!"
    echo
    log_info "Next steps:"
    log_info "1. Run: bash ./scripts/deploy.sh"
    log_info "2. Set up monitoring: bash ./scripts/setup-monitoring.sh"
    echo
}

# Run the deployment
main "$@"
