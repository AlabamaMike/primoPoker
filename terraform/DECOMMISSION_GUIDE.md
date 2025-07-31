# PrimoPoker Infrastructure Decommission Guide

## 🚨 URGENT: Cost-Saving Infrastructure Destruction

This guide will help you safely destroy your PrimoPoker production infrastructure to **stop incurring hosting charges**.

## 💰 Current Infrastructure & Estimated Monthly Costs

Based on your Terraform state, you have the following **expensive resources** running:

### High-Cost Resources (Primary Targets)
- **PostgreSQL Database Instance** (`db-f1-micro` in `us-west2`)
  - Estimated cost: ~$15-25/month
  - 💾 **Contains all game data, user accounts, hand history**
  
- **Redis Instance** (`BASIC` tier, 1GB in `us-west2`) 
  - Estimated cost: ~$25-35/month
  - 🔄 **Contains session data and real-time game state**

- **VPC Access Connector**
  - Estimated cost: ~$8-12/month
  - 🌐 **Network connectivity for Cloud Run**

### Medium-Cost Resources
- **VPC Network & Subnets**
  - Estimated cost: ~$2-5/month
  - 🔗 **Private networking infrastructure**

- **Storage Buckets**
  - Estimated cost: ~$1-3/month (depending on usage)
  - 📁 **Static assets and Terraform state**

- **Pub/Sub Topics & Subscriptions**
  - Estimated cost: ~$0-2/month (depending on message volume)
  - 📨 **Real-time game event messaging**

### **Total Estimated Monthly Savings: $50-80/month** 🎯

## 🔐 Prerequisites

1. **Google Cloud CLI installed and authenticated**
   ```bash
   gcloud auth login
   gcloud config set project primopoker
   ```

2. **Terraform installed** (v1.0+)
   ```bash
   terraform --version
   ```

3. **Access to the `primopoker` GCP project**

## ⚡ Quick Destruction (Automated)

### Option 1: Use the Automated Script (Recommended)

```bash
cd /workspaces/primoPoker/terraform
./decommission.sh
```

The script will:
- ✅ Prompt for confirmation (requires typing "DESTROY")
- ✅ Offer to backup database first
- ✅ Authenticate with GCP
- ✅ Create and execute destruction plan
- ✅ Clean up local state files

## 🛠️ Manual Destruction (Step-by-Step)

### Option 2: Manual Terraform Commands

If you prefer manual control:

```bash
# 1. Navigate to terraform directory
cd /workspaces/primoPoker/terraform

# 2. Authenticate with Google Cloud
gcloud auth application-default login
gcloud config set project primopoker

# 3. Initialize Terraform
terraform init

# 4. Create destruction plan
terraform plan -destroy

# 5. Review the plan carefully, then apply
terraform destroy
```

When prompted, type `yes` to confirm destruction.

## 📋 Resources That Will Be Destroyed

```
✅ google_sql_database_instance.postgres_instance
✅ google_redis_instance.redis_instance  
✅ google_vpc_access_connector.connector
✅ google_compute_network.vpc_network
✅ google_compute_subnetwork.app_subnet
✅ google_storage_bucket.static_assets
✅ google_storage_bucket.terraform_state
✅ google_pubsub_topic.game_updates
✅ google_pubsub_subscription.game_updates_subscription
✅ google_secret_manager_secret.db_password
✅ google_secret_manager_secret.jwt_secret
✅ All IAM roles and service accounts
✅ All enabled GCP APIs
```

## ⚠️ Data Backup (Optional but Recommended)

### Database Backup
```bash
# Export database before destruction
gcloud sql export sql primopoker-db-prod \
    gs://primopoker-backups/final-backup-$(date +%Y%m%d).sql \
    --database=primopoker
```

### Application Logs Backup
```bash
# Export recent logs
gcloud logging read 'resource.type="cloud_run_revision"' \
    --limit=5000 \
    --format=json > primopoker-logs-$(date +%Y%m%d).json
```

## 🔍 Verification Steps

After destruction, verify everything is gone:

1. **Check GCP Console**
   - Visit [GCP Console](https://console.cloud.google.com)
   - Navigate to each service (SQL, Redis, VPC, etc.)
   - Confirm no resources remain

2. **Check Billing**
   - Go to [GCP Billing](https://console.cloud.google.com/billing)
   - Monitor usage for next few days
   - Should see dramatic cost reduction

3. **Terraform State Check**
   ```bash
   terraform state list
   # Should return empty or minimal results
   ```

## 🚨 Emergency Manual Cleanup

If Terraform fails, manually delete via GCP Console:

1. **Cloud SQL** → Delete `primopoker-db-prod`
2. **Memory Store (Redis)** → Delete Redis instance
3. **VPC Networks** → Delete custom VPC
4. **Cloud Storage** → Delete all buckets
5. **Pub/Sub** → Delete topics and subscriptions

## 💡 Cost Monitoring

After destruction:
- Set up billing alerts for the project
- Consider deleting the entire GCP project if unused
- Monitor billing dashboard for 2-3 days to confirm charges stop

## 🎯 Expected Results

**Immediate:**
- All infrastructure destroyed
- No more compute/database charges
- Storage charges eliminated

**Within 24-48 hours:**
- Billing dashboard shows reduced usage
- Monthly cost projection drops to near $0

**Monthly Savings: $50-80** 💰

---

## 📞 Support

If you encounter issues:
1. Check GCP Console for any remaining resources
2. Review Terraform state: `terraform state list`
3. Manually delete stubborn resources via GCP Console
4. Contact GCP Support if needed

**Remember: The goal is to stop all recurring charges ASAP!** ⏰
