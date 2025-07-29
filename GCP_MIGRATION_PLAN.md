# 🚀 **Google Cloud Platform Migration Plan**
## **PrimoPoker Server - Cloud-Native Transformation**

---

## **📋 Executive Summary**

This document outlines the complete migration plan to transform the PrimoPoker server into a cloud-native application optimized for Google Cloud Platform (GCP). The migration focuses on scalability, reliability, security, and cost optimization.

---

## **🎯 Phase 1: Infrastructure & Containerization**

### **✅ Completed Components**

#### **1.1 Docker Containerization**
- ✅ **Dockerfile**: Multi-stage build with Alpine Linux base
- ✅ **Health checks**: Built-in HTTP health endpoint monitoring
- ✅ **Security**: Non-root user, minimal attack surface
- ✅ **Optimization**: Small image size (~20MB final image)

#### **1.2 Cloud Build Pipeline**
- ✅ **cloudbuild.yaml**: Automated CI/CD with Cloud Build
- ✅ **Container Registry**: Automatic image storage and versioning
- ✅ **Deployment**: Automated Cloud Run deployment

#### **1.3 Cloud Run Configuration**
- ✅ **service.yaml**: Declarative Cloud Run service configuration
- ✅ **Scaling**: Auto-scaling from 1-100 instances
- ✅ **Resources**: 1 CPU, 1Gi memory per instance
- ✅ **Concurrency**: 1000 concurrent requests per instance

---

## **🛠️ Phase 2: Cloud Services Integration**

### **✅ Completed Components**

#### **2.1 Secret Management**
- ✅ **Secret Manager Integration**: `internal/gcp/secrets.go`
- ✅ **Environment Variables**: Secure secret injection via Cloud Run
- ✅ **JWT Secrets**: Database passwords, API keys stored securely

#### **2.2 Cloud Logging**
- ✅ **Structured Logging**: `internal/gcp/logging.go`
- ✅ **Logrus Integration**: Cloud Logging hook for existing logs
- ✅ **Trace Correlation**: Request tracing and monitoring

#### **2.3 Configuration Updates**
- ✅ **GCP Config**: Extended configuration for cloud services
- ✅ **Environment Detection**: Development vs Production settings
- ✅ **Cloud SQL Support**: Unix socket and TCP connections

---

## **🗄️ Phase 3: Database & Storage**

### **Required Changes**

#### **3.1 Cloud SQL PostgreSQL**
```yaml
# Cloud SQL Instance Configuration
instance_name: primopoker-prod
database_version: POSTGRES_15
tier: db-custom-2-4096  # 2 vCPUs, 4GB RAM
storage_size: 100GB
storage_auto_resize: true
backup_enabled: true
point_in_time_recovery: true
high_availability: true
```

#### **3.2 Database Connection Updates**
```go
// Update database.go for Cloud SQL
func NewCloudSQLDB(config Config) (*DB, error) {
    var dsn string
    if config.SocketPath != "" {
        // Unix socket connection for Cloud SQL
        dsn = fmt.Sprintf(
            "host=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=%s",
            config.SocketPath, config.User, config.Password, 
            config.DBName, config.TimeZone,
        )
    } else {
        // TCP connection
        dsn = fmt.Sprintf(
            "host=%s port=%d user=%s password=%s dbname=%s sslmode=require TimeZone=%s",
            config.Host, config.Port, config.User, config.Password,
            config.DBName, config.TimeZone,
        )
    }
    // ... rest of connection logic
}
```

#### **3.3 Connection Pooling Optimization**
- **Max Connections**: 25 per instance (Cloud SQL limit: 100)
- **Idle Timeout**: 10 minutes
- **Connection Lifetime**: 1 hour
- **Retry Logic**: Exponential backoff for connection failures

---

## **🔄 Phase 4: Real-time Communication**

### **4.1 WebSocket Scalability Challenge**
**Current Issue**: WebSockets don't work well with Cloud Run's stateless model

**Solution Options**:

#### **Option A: Pub/Sub + Server-Sent Events (Recommended)**
```go
// Replace WebSocket with Pub/Sub + SSE
type RealtimeService struct {
    pubsubClient *pubsub.Client
    topic        *pubsub.Topic
}

func (rs *RealtimeService) BroadcastGameUpdate(gameID string, update GameUpdate) {
    message := &pubsub.Message{
        Data: json.Marshal(update),
        Attributes: map[string]string{
            "gameId": gameID,
            "type":   "game_update",
        },
    }
    rs.topic.Publish(ctx, message)
}
```

#### **Option B: Redis Pub/Sub with Memorystore**
```go
// Use Redis for real-time messaging
type RedisRealtimeService struct {
    client *redis.Client
}

func (rs *RedisRealtimeService) Subscribe(gameID string) *redis.PubSub {
    return rs.client.Subscribe(ctx, fmt.Sprintf("game:%s", gameID))
}
```

#### **Option C: Firebase Realtime Database**
- **Pros**: Built-in WebSocket handling, automatic scaling
- **Cons**: Additional service dependency, data model changes

---

## **📊 Phase 5: Monitoring & Observability**

### **5.1 Cloud Monitoring Integration**
```go
// Custom metrics for poker-specific monitoring
type PokerMetrics struct {
    gamesActive      monitoring.Gauge
    playersOnline    monitoring.Gauge
    handsPerMinute   monitoring.Counter
    responseLatency  monitoring.Histogram
}

func (pm *PokerMetrics) RecordGameStart(gameID string) {
    pm.gamesActive.Set(getCurrentActiveGames())
}
```

### **5.2 Error Reporting**
```go
import "cloud.google.com/go/errorreporting"

func setupErrorReporting(projectID string) *errorreporting.Client {
    client, err := errorreporting.NewClient(ctx, projectID, errorreporting.Config{
        ServiceName:    "primopoker",
        ServiceVersion: getVersion(),
    })
    return client
}
```

### **5.3 Performance Monitoring**
- **Cloud Trace**: Request tracing and latency analysis
- **Cloud Profiler**: CPU and memory profiling
- **Custom Dashboards**: Poker-specific metrics visualization

---

## **🔐 Phase 6: Security & Compliance**

### **6.1 Identity and Access Management**
```yaml
# IAM roles for the application
roles:
  - role: roles/cloudsql.client
    members: ["serviceAccount:primopoker@project.iam.gserviceaccount.com"]
  - role: roles/secretmanager.secretAccessor
    members: ["serviceAccount:primopoker@project.iam.gserviceaccount.com"]
  - role: roles/pubsub.publisher
    members: ["serviceAccount:primopoker@project.iam.gserviceaccount.com"]
```

### **6.2 Network Security**
- **Cloud Armor**: DDoS protection and WAF rules
- **VPC Peering**: Secure database connections
- **SSL/TLS**: End-to-end encryption

### **6.3 Audit Logging**
```go
func auditLog(action string, userID string, details map[string]interface{}) {
    entry := LogEntry{
        Severity: "info",
        Message:  fmt.Sprintf("Audit: %s", action),
        Labels: map[string]string{
            "audit":   "true",
            "user_id": userID,
            "action":  action,
        },
        Data: details,
    }
    cloudLogger.Log(entry)
}
```

---

## **🚀 Phase 7: Deployment & Scaling**

### **7.1 Multi-Region Deployment**
```yaml
# Cloud Run services in multiple regions
regions:
  - us-central1    # Primary
  - us-east1       # Secondary
  - europe-west1   # EU users
  - asia-east1     # Asian users
```

### **7.2 Load Balancing**
```yaml
# Global Load Balancer configuration
backend_services:
  - name: primopoker-backend
    backends:
      - group: us-central1-primopoker
        balancing_mode: UTILIZATION
        max_utilization: 0.8
      - group: us-east1-primopoker
        balancing_mode: UTILIZATION
        max_utilization: 0.8
```

### **7.3 Auto-scaling Configuration**
```yaml
# Cloud Run scaling settings
scaling:
  min_instances: 2      # Always warm instances
  max_instances: 100    # Scale up to 100 instances
  target_cpu: 70        # Scale at 70% CPU utilization
  target_concurrency: 800  # Scale at 800 concurrent requests
```

---

## **💰 Phase 8: Cost Optimization**

### **8.1 Resource Right-sizing**
| Component | Current | Optimized | Monthly Cost |
|-----------|---------|-----------|--------------|
| Cloud Run | 1 CPU, 1GB RAM | 0.5 CPU, 512MB RAM | $50-200 |
| Cloud SQL | db-custom-2-4096 | db-custom-1-3840 | $150-300 |
| Memorystore | 1GB Redis | 1GB Redis | $50 |
| **Total** | | | **$250-550** |

### **8.2 Cost Monitoring**
```go
// Cost allocation labels
labels := map[string]string{
    "environment": "production",
    "service":     "primopoker",
    "component":   "game-server",
    "team":        "backend",
}
```

---

## **📈 Phase 9: Performance Optimization**

### **9.1 Caching Strategy**
```go
// Redis caching for frequently accessed data
type CacheService struct {
    redis *redis.Client
}

func (cs *CacheService) CacheGameState(gameID string, state GameState) {
    data, _ := json.Marshal(state)
    cs.redis.Set(ctx, fmt.Sprintf("game:%s", gameID), data, 10*time.Minute)
}
```

### **9.2 Database Optimization**
```sql
-- Indexes for optimal query performance
CREATE INDEX CONCURRENTLY idx_hand_history_user_time 
ON hand_history(user_id, started_at DESC);

CREATE INDEX CONCURRENTLY idx_games_active 
ON games(status, created_at DESC) 
WHERE status = 'active';

CREATE INDEX CONCURRENTLY idx_game_participants 
ON game_participations(game_id, user_id);
```

### **9.3 Connection Pooling**
```go
// Optimized connection pool settings
config := database.Config{
    MaxOpenConns:    25,  // Cloud SQL connection limit
    MaxIdleConns:    5,   // Keep some connections warm
    ConnMaxLifetime: time.Hour,
    ConnMaxIdleTime: 10 * time.Minute,
}
```

---

## **🧪 Phase 10: Testing & Validation**

### **10.1 Load Testing**
```yaml
# Load testing configuration
load_test:
  concurrent_users: 1000
  test_duration: 30m
  scenarios:
    - name: "user_registration"
      weight: 10%
    - name: "game_creation"
      weight: 20%
    - name: "poker_gameplay"
      weight: 60%
    - name: "metrics_viewing"
      weight: 10%
```

### **10.2 Disaster Recovery Testing**
- **Database Failover**: Test Cloud SQL high availability
- **Regional Outage**: Test multi-region deployment
- **Secret Rotation**: Test Secret Manager integration

---

## **📅 Implementation Timeline**

### **Week 1-2: Infrastructure Setup**
- [ ] Create GCP project and enable APIs
- [ ] Set up Cloud SQL instance with high availability
- [ ] Configure Secret Manager with all secrets
- [ ] Set up Cloud Build pipeline

### **Week 3-4: Application Updates**
- [ ] Implement Secret Manager integration
- [ ] Update database connection for Cloud SQL
- [ ] Add Cloud Logging integration
- [ ] Update configuration management

### **Week 5-6: Real-time Communication**
- [ ] Implement Pub/Sub + SSE solution
- [ ] Set up Memorystore Redis instance
- [ ] Update WebSocket handlers
- [ ] Test real-time functionality

### **Week 7-8: Monitoring & Security**
- [ ] Set up Cloud Monitoring dashboards
- [ ] Implement error reporting
- [ ] Configure Cloud Armor protection
- [ ] Set up audit logging

### **Week 9-10: Testing & Optimization**
- [ ] Load testing and performance optimization
- [ ] Security testing and vulnerability assessment
- [ ] Disaster recovery testing
- [ ] Cost optimization analysis

### **Week 11-12: Deployment & Go-Live**
- [ ] Blue-green deployment setup
- [ ] Production deployment
- [ ] Monitoring and alerting validation
- [ ] Post-deployment optimization

---

## **🎯 Success Metrics**

### **Performance Targets**
- **Response Time**: < 100ms for API endpoints
- **WebSocket Latency**: < 50ms for real-time updates
- **Availability**: 99.9% uptime SLA
- **Scalability**: Handle 10,000+ concurrent users

### **Cost Targets**
- **Monthly Cost**: < $500 for moderate traffic
- **Cost per User**: < $0.50 per active user per month
- **Database Cost**: < 30% of total infrastructure cost

### **Security Targets**
- **Zero** exposed secrets in code or configuration
- **100%** encrypted data in transit and at rest
- **SOC 2 Type II** compliance ready
- **GDPR** compliance for EU users

---

## **🚨 Risks & Mitigations**

### **High-Risk Items**
1. **WebSocket Migration**: Complex transition from stateful to stateless
   - *Mitigation*: Implement Pub/Sub + SSE with backward compatibility

2. **Database Migration**: Potential downtime during migration
   - *Mitigation*: Use read replicas and blue-green deployment

3. **Real-time Performance**: Latency increase with cloud services
   - *Mitigation*: Implement caching and optimize database queries

### **Medium-Risk Items**
1. **Cost Overruns**: Cloud costs can spiral quickly
   - *Mitigation*: Implement cost alerts and regular optimization reviews

2. **Vendor Lock-in**: Heavy reliance on GCP services
   - *Mitigation*: Use standard interfaces and maintain portability

---

## **✅ Next Steps**

1. **Review and Approve Plan**: Stakeholder sign-off on migration approach
2. **Create GCP Project**: Set up billing, IAM, and initial resources
3. **Development Environment**: Create staging environment for testing
4. **Begin Phase 1**: Start with containerization and basic cloud services
5. **Iterative Testing**: Continuous testing throughout migration process

---

**📞 For questions or clarifications, contact the DevOps team.**

*Last Updated: July 29, 2025*
