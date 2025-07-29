#!/bin/bash

# 🚀 PrimoPoker GCP Deployment Script
# This script automates the complete deployment to Google Cloud Platform

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

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if required tools are installed
    command -v gcloud >/dev/null 2>&1 || log_error "gcloud CLI is not installed"
    command -v terraform >/dev/null 2>&1 || log_error "Terraform is not installed"
    command -v docker >/dev/null 2>&1 || log_error "Docker is not installed"
    
    # Check if PROJECT_ID is set
    if [ -z "$PROJECT_ID" ]; then
        log_error "PROJECT_ID environment variable is not set"
    fi
    
    # Check if authenticated with gcloud
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
        log_error "Not authenticated with gcloud. Run 'gcloud auth login'"
    fi
    
    # Check if terraform.tfvars exists
    if [ ! -f "$TERRAFORM_DIR/terraform.tfvars" ]; then
        log_error "terraform.tfvars not found. Copy terraform.tfvars.example and fill in your values"
    fi
    
    log_success "Prerequisites check passed"
}

setup_gcp_project() {
    log_info "Setting up GCP project..."
    
    # Set the project
    gcloud config set project "$PROJECT_ID"
    
    # Enable billing (this needs to be done manually in the console)
    log_warning "Ensure billing is enabled for project $PROJECT_ID"
    
    log_success "GCP project setup completed"
}

deploy_infrastructure() {
    log_info "Deploying infrastructure with Terraform..."
    
    cd "$TERRAFORM_DIR"
    
    # Initialize Terraform
    terraform init
    
    # Plan the deployment
    terraform plan -var="project_id=$PROJECT_ID" -var="region=$REGION"
    
    # Ask for confirmation
    read -p "Do you want to proceed with the infrastructure deployment? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_error "Deployment cancelled by user"
    fi
    
    # Apply the infrastructure
    terraform apply -auto-approve -var="project_id=$PROJECT_ID" -var="region=$REGION"
    
    cd ..
    log_success "Infrastructure deployment completed"
}

build_and_push_image() {
    log_info "Building and pushing Docker image..."
    
    # Build the image
    docker build -t gcr.io/$PROJECT_ID/$SERVICE_NAME:latest .
    
    # Configure Docker to use gcloud as a credential helper
    gcloud auth configure-docker
    
    # Push the image
    docker push gcr.io/$PROJECT_ID/$SERVICE_NAME:latest
    
    log_success "Docker image built and pushed"
}

deploy_application() {
    log_info "Deploying application to Cloud Run..."
    
    # Get database connection name from Terraform output
    DB_CONNECTION_NAME=$(cd "$TERRAFORM_DIR" && terraform output -raw database_connection_name)
    SERVICE_ACCOUNT_EMAIL=$(cd "$TERRAFORM_DIR" && terraform output -raw service_account_email)
    VPC_CONNECTOR=$(cd "$TERRAFORM_DIR" && terraform output -raw vpc_connector_name)
    
    # Deploy to Cloud Run
    gcloud run deploy $SERVICE_NAME \
        --image gcr.io/$PROJECT_ID/$SERVICE_NAME:latest \
        --platform managed \
        --region $REGION \
        --service-account $SERVICE_ACCOUNT_EMAIL \
        --vpc-connector $VPC_CONNECTOR \
        --set-env-vars "PROJECT_ID=$PROJECT_ID" \
        --set-env-vars "DB_CONNECTION_NAME=$DB_CONNECTION_NAME" \
        --set-env-vars "ENVIRONMENT=production" \
        --set-env-vars "REGION=$REGION" \
        --allow-unauthenticated \
        --memory 1Gi \
        --cpu 1 \
        --concurrency 1000 \
        --max-instances 100 \
        --min-instances 2 \
        --port 8080 \
        --timeout 300 \
        --no-traffic
    
    log_success "Application deployed to Cloud Run"
}

setup_monitoring() {
    log_info "Setting up monitoring and alerting..."
    
    # Create uptime check
    gcloud alpha monitoring uptime create \
        --display-name="$SERVICE_NAME Uptime Check" \
        --http-check-path="/health" \
        --hostname="$SERVICE_NAME-$PROJECT_ID.a.run.app" \
        --project="$PROJECT_ID"
    
    log_success "Monitoring setup completed"
}

run_health_check() {
    log_info "Running health check..."
    
    # Get the service URL
    SERVICE_URL=$(gcloud run services describe $SERVICE_NAME --platform managed --region $REGION --format 'value(status.url)')
    
    # Wait for service to be ready
    sleep 30
    
    # Check health endpoint
    if curl -f "$SERVICE_URL/health" > /dev/null 2>&1; then
        log_success "Health check passed - service is running"
    else
        log_error "Health check failed - service may not be running correctly"
    fi
    
    echo
    log_success "Deployment completed successfully!"
    log_info "Service URL: $SERVICE_URL"
    log_info "You can view logs with: gcloud logging read 'resource.type=cloud_run_revision AND resource.labels.service_name=$SERVICE_NAME' --limit 50 --format json"
}

migrate_traffic() {
    log_info "Migrating traffic to new revision..."
    
    read -p "Do you want to migrate 100% traffic to the new revision? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        gcloud run services update-traffic $SERVICE_NAME \
            --to-latest \
            --platform managed \
            --region $REGION
        log_success "Traffic migrated to new revision"
    else
        log_info "Traffic migration skipped. Manually update traffic allocation when ready."
    fi
}

# Main execution
main() {
    echo
    log_info "🚀 Starting PrimoPoker GCP Deployment"
    echo
    
    check_prerequisites
    setup_gcp_project
    deploy_infrastructure
    build_and_push_image
    deploy_application
    setup_monitoring
    run_health_check
    migrate_traffic
    
    echo
    log_success "🎉 Deployment completed successfully!"
    echo
    log_info "Next steps:"
    log_info "1. Test your application thoroughly"
    log_info "2. Set up custom domain (if needed)"
    log_info "3. Configure monitoring alerts"
    log_info "4. Set up CI/CD pipeline"
    echo
}

# Run the deployment
main "$@"
