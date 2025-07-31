#!/bin/bash

# PrimoPoker Infrastructure Decommission Script
# This script safely destroys all GCP resources to avoid hosting charges

set -e

echo "🚨 PrimoPoker Infrastructure Decommission"
echo "=========================================="
echo ""
echo "⚠️  WARNING: This will destroy ALL production infrastructure!"
echo "    - PostgreSQL database (all data will be lost)"
echo "    - Redis cache"
echo "    - Storage buckets"
echo "    - VPC and networking"
echo "    - All GCP services and APIs"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to prompt for confirmation
confirm_destruction() {
    echo -e "${RED}Are you absolutely sure you want to destroy the production environment?${NC}"
    echo "This action cannot be undone and will result in complete data loss."
    echo ""
    echo "Type 'DESTROY' (in capital letters) to confirm:"
    read -r confirmation
    
    if [ "$confirmation" != "DESTROY" ]; then
        echo "❌ Destruction cancelled. Infrastructure preserved."
        exit 0
    fi
}

# Function to backup important data (optional)
backup_data() {
    log_warning "Consider backing up important data before destruction!"
    echo ""
    echo "Backup commands you can run manually:"
    echo "1. Database backup:"
    echo "   gcloud sql export sql primopoker-db-prod gs://your-backup-bucket/db-backup-\$(date +%Y%m%d).sql"
    echo ""
    echo "2. Download logs:"
    echo "   gcloud logging read 'resource.type=\"cloud_run_revision\"' --limit=1000 --format=json > logs-backup-\$(date +%Y%m%d).json"
    echo ""
    echo "Do you want to continue without backing up? (y/N):"
    read -r skip_backup
    
    if [[ ! "$skip_backup" =~ ^[Yy]$ ]]; then
        echo "❌ Destruction cancelled. Please backup your data first."
        exit 0
    fi
}

# Main destruction process
main() {
    log_info "Starting PrimoPoker infrastructure decommission..."
    
    # Check if we're in the right directory
    if [ ! -f "terraform.tfvars" ] || [ ! -f "main.tf" ]; then
        log_error "Please run this script from the terraform/ directory"
        exit 1
    fi
    
    # Prompt for confirmation
    confirm_destruction
    
    # Ask about backup
    backup_data
    
    log_info "Step 1: Authenticating with Google Cloud..."
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | head -1; then
        log_warning "Not authenticated with gcloud. Running authentication..."
        gcloud auth login
    fi
    
    # Set the project
    PROJECT_ID=$(grep 'project_id' terraform.tfvars | cut -d '"' -f 2)
    log_info "Setting project to: $PROJECT_ID"
    gcloud config set project "$PROJECT_ID"
    
    log_info "Step 2: Initializing Terraform..."
    terraform init
    
    log_info "Step 3: Creating destruction plan..."
    terraform plan -destroy -out=destroy.tfplan
    
    log_warning "Please review the destruction plan above carefully!"
    echo "Press Enter to continue with destruction, or Ctrl+C to cancel:"
    read -r
    
    log_info "Step 4: Executing infrastructure destruction..."
    terraform apply destroy.tfplan
    
    log_info "Step 5: Cleaning up local state..."
    rm -f destroy.tfplan
    rm -f terraform.tfstate.backup.*
    
    # Optional: Remove the entire .terraform directory
    echo "Do you want to remove Terraform cache and state files? (y/N):"
    read -r clean_terraform
    
    if [[ "$clean_terraform" =~ ^[Yy]$ ]]; then
        rm -rf .terraform/
        rm -f .terraform.lock.hcl
        log_success "Terraform cache cleaned"
    fi
    
    log_success "🎉 Infrastructure successfully decommissioned!"
    echo ""
    echo "💰 Cost Summary:"
    echo "   ✅ PostgreSQL database instance: DESTROYED"
    echo "   ✅ Redis instance: DESTROYED" 
    echo "   ✅ VPC and networking: DESTROYED"
    echo "   ✅ Storage buckets: DESTROYED"
    echo "   ✅ All GCP APIs: Disabled"
    echo ""
    echo "You should no longer incur hosting charges for PrimoPoker infrastructure."
    echo ""
    echo "📋 Next steps:"
    echo "   1. Verify in GCP Console that all resources are gone"
    echo "   2. Check GCP billing to confirm charges have stopped"
    echo "   3. Consider disabling the GCP project entirely if not needed"
    echo ""
    log_info "Decommission completed successfully!"
}

# Run main function
main "$@"
