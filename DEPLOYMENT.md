# 🚀 **PrimoPoker GCP Deployment Guide**

This guide walks you through deploying PrimoPoker to Google Cloud Platform using our cloud-native architecture.

## **📋 Prerequisites**

### **Required Tools**
- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install) (gcloud CLI)
- [Terraform](https://www.terraform.io/downloads) (>= 1.0)
- [Docker](https://docs.docker.com/get-docker/)
- [Git](https://git-scm.com/downloads)

### **GCP Setup**
1. Create a new GCP project or use an existing one
2. Enable billing for the project
3. Install and authenticate with gcloud:
   ```bash
   gcloud auth login
   gcloud config set project YOUR_PROJECT_ID
   ```

## **🛠️ Quick Deployment**

### **1. Clone and Configure**
```bash
git clone <your-repo-url>
cd primoPoker

# Copy and configure Terraform variables
cp terraform/terraform.tfvars.example terraform/terraform.tfvars
# Edit terraform/terraform.tfvars with your values
```

### **2. Set Environment Variables**
```bash
export PROJECT_ID="your-gcp-project-id"
export REGION="us-central1"  # Optional, defaults to us-central1
export NOTIFICATION_EMAIL="alerts@yourcompany.com"  # Optional
```

### **3. Run Automated Deployment**
```bash
chmod +x scripts/deploy.sh
./scripts/deploy.sh
```

### **4. Set Up Monitoring**
```bash
chmod +x scripts/setup-monitoring.sh
./scripts/setup-monitoring.sh
```

## **📊 Manual Deployment Steps**

If you prefer to deploy manually or understand each step:

### **Step 1: Infrastructure Deployment**
```bash
cd terraform
terraform init
terraform plan -var="project_id=$PROJECT_ID"
terraform apply -var="project_id=$PROJECT_ID"
cd ..
```

### **Step 2: Build and Push Container**
```bash
# Build the Docker image
docker build -t gcr.io/$PROJECT_ID/primopoker:latest .

# Configure Docker for GCP
gcloud auth configure-docker

# Push the image
docker push gcr.io/$PROJECT_ID/primopoker:latest
```

### **Step 3: Deploy to Cloud Run**
```bash
# Get infrastructure outputs
DB_CONNECTION_NAME=$(cd terraform && terraform output -raw database_connection_name)
SERVICE_ACCOUNT_EMAIL=$(cd terraform && terraform output -raw service_account_email)
VPC_CONNECTOR=$(cd terraform && terraform output -raw vpc_connector_name)

# Deploy to Cloud Run
gcloud run deploy primopoker \
    --image gcr.io/$PROJECT_ID/primopoker:latest \
    --platform managed \
    --region $REGION \
    --service-account $SERVICE_ACCOUNT_EMAIL \
    --vpc-connector $VPC_CONNECTOR \
    --set-env-vars "PROJECT_ID=$PROJECT_ID,DB_CONNECTION_NAME=$DB_CONNECTION_NAME,ENVIRONMENT=production" \
    --allow-unauthenticated \
    --memory 1Gi \
    --cpu 1 \
    --concurrency 1000 \
    --max-instances 100 \
    --min-instances 2
```

## **⚙️ Configuration**

### **Environment Variables**
The application uses the following environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `PROJECT_ID` | GCP Project ID | `my-poker-project` |
| `DB_CONNECTION_NAME` | Cloud SQL connection name | `project:region:instance` |
| `ENVIRONMENT` | Deployment environment | `production` |
| `REGION` | GCP region | `us-central1` |
| `JWT_SECRET_NAME` | Secret Manager secret name | `primopoker-jwt-secret` |
| `DB_PASSWORD_SECRET_NAME` | Database password secret name | `primopoker-db-password` |

### **Secrets Configuration**
Secrets are managed via Google Secret Manager:

1. **JWT Secret**: Used for authentication tokens
2. **Database Password**: PostgreSQL user password
3. **Additional secrets**: Can be added as needed

## **🏗️ Architecture Overview**

### **Components Deployed**
- **Cloud Run**: Serverless application hosting
- **Cloud SQL**: PostgreSQL database with high availability
- **Memorystore**: Redis for caching and real-time features
- **Pub/Sub**: Message queue for game events
- **Secret Manager**: Secure secret storage
- **VPC**: Private networking
- **Cloud Armor**: Security and DDoS protection
- **Cloud Monitoring**: Observability and alerting

### **Network Architecture**
```
Internet → Cloud Armor → Load Balancer → Cloud Run
                                           ↓
                                      VPC Connector
                                           ↓
                     Private Network (Cloud SQL, Redis)
```

## **📊 Monitoring & Operations**

### **Key Metrics**
- **Request Rate**: Requests per second
- **Response Latency**: P50, P95, P99 latencies
- **Error Rate**: 4xx and 5xx error percentages
- **Active Instances**: Auto-scaling metrics
- **Database Connections**: Connection pool usage

### **Alerts Configured**
- High error rate (>5%)
- High latency (>1 second)
- Database connection issues
- Service downtime

### **Dashboards**
Access monitoring dashboards at:
```
https://console.cloud.google.com/monitoring/dashboards?project=YOUR_PROJECT_ID
```

## **🔐 Security Features**

### **Network Security**
- Private VPC with no public IP addresses
- Cloud Armor protection against DDoS
- SSL/TLS encryption for all traffic

### **Application Security**
- No secrets in code or environment variables
- Service account with minimal required permissions
- Encrypted secrets in Secret Manager

### **Database Security**
- Private IP only (no public access)
- Encrypted at rest and in transit
- Automated backups with point-in-time recovery

## **💰 Cost Optimization**

### **Estimated Monthly Costs** (moderate traffic)
- Cloud Run: $50-200
- Cloud SQL: $150-300
- Memorystore: $50
- Other services: $50-100
- **Total**: $300-650/month

### **Cost Optimization Tips**
1. **Right-size resources**: Monitor and adjust CPU/Memory
2. **Use preemptible instances**: For development environments
3. **Set up cost alerts**: Monitor spending patterns
4. **Regular reviews**: Monthly cost optimization reviews

## **🚨 Troubleshooting**

### **Common Issues**

#### **Deployment Fails**
```bash
# Check Cloud Build logs
gcloud builds log --region=$REGION

# Check Cloud Run deployment
gcloud run services describe primopoker --region=$REGION
```

#### **Database Connection Issues**
```bash
# Test database connectivity
gcloud sql connect INSTANCE_NAME --user=primopoker_user

# Check database logs
gcloud logging read "resource.type=cloudsql_database" --limit=50
```

#### **High Memory Usage**
```bash
# Check instance metrics
gcloud run services describe primopoker --region=$REGION --format="value(status.traffic)"

# Scale up resources if needed
gcloud run services update primopoker --memory=2Gi --region=$REGION
```

### **Debug Commands**
```bash
# View application logs
gcloud logging read 'resource.type=cloud_run_revision AND resource.labels.service_name=primopoker' --limit=50

# Check service health
curl https://primopoker-PROJECT_ID.a.run.app/health

# Monitor real-time logs
gcloud logging tail 'resource.type=cloud_run_revision'
```

## **🔄 CI/CD Pipeline**

### **Cloud Build Configuration**
The `cloudbuild.yaml` file sets up automated deployments:

1. **Triggers**: On push to main branch
2. **Build**: Docker image creation
3. **Test**: Run unit tests
4. **Deploy**: Update Cloud Run service
5. **Notify**: Send deployment notifications

### **Manual Trigger**
```bash
gcloud builds submit --config cloudbuild.yaml .
```

## **📈 Scaling Considerations**

### **Horizontal Scaling**
- Cloud Run auto-scales from 2-100 instances
- Database connection pooling handles concurrent connections
- Redis handles caching for frequently accessed data

### **Database Scaling**
```bash
# Scale up database instance
gcloud sql instances patch INSTANCE_NAME --tier=db-custom-4-8192

# Add read replicas for read-heavy workloads
gcloud sql instances create read-replica --master-instance-name=INSTANCE_NAME
```

## **🔄 Updates & Maintenance**

### **Application Updates**
```bash
# Build and deploy new version
docker build -t gcr.io/$PROJECT_ID/primopoker:v1.1.0 .
docker push gcr.io/$PROJECT_ID/primopoker:v1.1.0

# Deploy with zero downtime
gcloud run deploy primopoker --image gcr.io/$PROJECT_ID/primopoker:v1.1.0 --region=$REGION
```

### **Database Maintenance**
- Automated backups run daily
- Point-in-time recovery available
- Maintenance windows automatically scheduled

## **📞 Support**

### **Getting Help**
1. Check application logs first
2. Review monitoring dashboards
3. Consult this documentation
4. Contact the DevOps team

### **Emergency Procedures**
- **Service Down**: Check Cloud Run service status
- **Database Issues**: Check Cloud SQL maintenance windows
- **High Traffic**: Monitor auto-scaling behavior

---

**🎉 Congratulations!** Your PrimoPoker server is now running on Google Cloud Platform with enterprise-grade reliability, security, and scalability.

*For additional support, please contact the development team.*
