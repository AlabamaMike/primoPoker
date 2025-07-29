#!/bin/bash

# 📊 PrimoPoker Monitoring Setup Script
# Sets up comprehensive monitoring, alerting, and dashboards for GCP deployment

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
NOTIFICATION_EMAIL="${NOTIFICATION_EMAIL:-}"

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
    
    if [ -z "$PROJECT_ID" ]; then
        log_error "PROJECT_ID environment variable is not set"
    fi
    
    if [ -z "$NOTIFICATION_EMAIL" ]; then
        log_warning "NOTIFICATION_EMAIL not set - alerts will not be sent"
    fi
    
    command -v gcloud >/dev/null 2>&1 || log_error "gcloud CLI is not installed"
    
    gcloud config set project "$PROJECT_ID"
    
    log_success "Prerequisites check passed"
}

create_notification_channel() {
    if [ -n "$NOTIFICATION_EMAIL" ]; then
        log_info "Creating notification channel..."
        
        cat > notification_channel.json << EOF
{
  "type": "email",
  "displayName": "PrimoPoker Alerts",
  "labels": {
    "email_address": "$NOTIFICATION_EMAIL"
  }
}
EOF
        
        NOTIFICATION_CHANNEL=$(gcloud alpha monitoring channels create --channel-content-from-file=notification_channel.json --format="value(name)")
        rm notification_channel.json
        
        log_success "Notification channel created: $NOTIFICATION_CHANNEL"
        echo "$NOTIFICATION_CHANNEL" > notification_channel_name.txt
    else
        log_warning "Skipping notification channel creation"
    fi
}

create_uptime_check() {
    log_info "Creating uptime check..."
    
    SERVICE_URL=$(gcloud run services describe $SERVICE_NAME --platform managed --region $REGION --format 'value(status.url)')
    HOSTNAME=$(echo $SERVICE_URL | sed 's|https://||' | sed 's|http://||')
    
    cat > uptime_check.json << EOF
{
  "displayName": "PrimoPoker Health Check",
  "httpCheck": {
    "path": "/health",
    "port": 443,
    "requestMethod": "GET",
    "useSsl": true
  },
  "monitoredResource": {
    "type": "uptime_url",
    "labels": {
      "project_id": "$PROJECT_ID",
      "host": "$HOSTNAME"
    }
  },
  "timeout": "10s",
  "period": "60s",
  "selectedRegions": [
    "USA_OREGON",
    "USA_IOWA",
    "EUROPE_IRELAND"
  ]
}
EOF
    
    UPTIME_CHECK=$(gcloud alpha monitoring uptime create --uptime-check-from-file=uptime_check.json --format="value(name)")
    rm uptime_check.json
    
    log_success "Uptime check created: $UPTIME_CHECK"
}

create_alert_policies() {
    log_info "Creating alert policies..."
    
    # High Error Rate Alert
    cat > high_error_rate_alert.json << EOF
{
  "displayName": "PrimoPoker - High Error Rate",
  "documentation": {
    "content": "Error rate is above 5% for the PrimoPoker service"
  },
  "conditions": [
    {
      "displayName": "Error rate condition",
      "conditionThreshold": {
        "filter": "resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"$SERVICE_NAME\" AND metric.type=\"run.googleapis.com/request_count\"",
        "aggregations": [
          {
            "alignmentPeriod": "300s",
            "perSeriesAligner": "ALIGN_RATE",
            "crossSeriesReducer": "REDUCE_SUM",
            "groupByFields": ["resource.labels.service_name"]
          }
        ],
        "comparison": "COMPARISON_GREATER_THAN",
        "thresholdValue": 0.05,
        "duration": "300s"
      }
    }
  ],
  "combiner": "OR",
  "enabled": true
}
EOF
    
    if [ -n "$NOTIFICATION_EMAIL" ] && [ -f notification_channel_name.txt ]; then
        NOTIFICATION_CHANNEL=$(cat notification_channel_name.txt)
        cat >> high_error_rate_alert.json << EOF
,
  "notificationChannels": ["$NOTIFICATION_CHANNEL"]
EOF
    fi
    
    echo "}" >> high_error_rate_alert.json
    
    gcloud alpha monitoring policies create --policy-from-file=high_error_rate_alert.json
    rm high_error_rate_alert.json
    
    # High Latency Alert
    cat > high_latency_alert.json << EOF
{
  "displayName": "PrimoPoker - High Latency",
  "documentation": {
    "content": "Response latency is above 1 second for the PrimoPoker service"
  },
  "conditions": [
    {
      "displayName": "Latency condition",
      "conditionThreshold": {
        "filter": "resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"$SERVICE_NAME\" AND metric.type=\"run.googleapis.com/request_latencies\"",
        "aggregations": [
          {
            "alignmentPeriod": "300s",
            "perSeriesAligner": "ALIGN_DELTA",
            "crossSeriesReducer": "REDUCE_MEAN",
            "groupByFields": ["resource.labels.service_name"]
          }
        ],
        "comparison": "COMPARISON_GREATER_THAN",
        "thresholdValue": 1000,
        "duration": "300s"
      }
    }
  ],
  "combiner": "OR",
  "enabled": true
}
EOF
    
    if [ -n "$NOTIFICATION_EMAIL" ] && [ -f notification_channel_name.txt ]; then
        NOTIFICATION_CHANNEL=$(cat notification_channel_name.txt)
        cat >> high_latency_alert.json << EOF
,
  "notificationChannels": ["$NOTIFICATION_CHANNEL"]
EOF
    fi
    
    echo "}" >> high_latency_alert.json
    
    gcloud alpha monitoring policies create --policy-from-file=high_latency_alert.json
    rm high_latency_alert.json
    
    # Database Connection Alert
    cat > db_connection_alert.json << EOF
{
  "displayName": "PrimoPoker - Database Connection Issues",
  "documentation": {
    "content": "Database connections are failing for the PrimoPoker service"
  },
  "conditions": [
    {
      "displayName": "Database connection condition",
      "conditionThreshold": {
        "filter": "resource.type=\"cloudsql_database\" AND metric.type=\"cloudsql.googleapis.com/database/postgresql/num_backends\"",
        "aggregations": [
          {
            "alignmentPeriod": "300s",
            "perSeriesAligner": "ALIGN_MEAN",
            "crossSeriesReducer": "REDUCE_MEAN"
          }
        ],
        "comparison": "COMPARISON_LESS_THAN",
        "thresholdValue": 1,
        "duration": "300s"
      }
    }
  ],
  "combiner": "OR",
  "enabled": true
}
EOF
    
    if [ -n "$NOTIFICATION_EMAIL" ] && [ -f notification_channel_name.txt ]; then
        NOTIFICATION_CHANNEL=$(cat notification_channel_name.txt)
        cat >> db_connection_alert.json << EOF
,
  "notificationChannels": ["$NOTIFICATION_CHANNEL"]
EOF
    fi
    
    echo "}" >> db_connection_alert.json
    
    gcloud alpha monitoring policies create --policy-from-file=db_connection_alert.json
    rm db_connection_alert.json
    
    log_success "Alert policies created"
}

create_dashboard() {
    log_info "Creating monitoring dashboard..."
    
    cat > dashboard.json << EOF
{
  "displayName": "PrimoPoker - Production Dashboard",
  "mosaicLayout": {
    "tiles": [
      {
        "width": 6,
        "height": 4,
        "widget": {
          "title": "Request Count",
          "xyChart": {
            "dataSets": [
              {
                "timeSeriesQuery": {
                  "timeSeriesFilter": {
                    "filter": "resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"$SERVICE_NAME\" AND metric.type=\"run.googleapis.com/request_count\"",
                    "aggregation": {
                      "alignmentPeriod": "60s",
                      "perSeriesAligner": "ALIGN_RATE",
                      "crossSeriesReducer": "REDUCE_SUM"
                    }
                  }
                },
                "plotType": "LINE"
              }
            ],
            "timeshiftDuration": "0s",
            "yAxis": {
              "label": "Requests/sec",
              "scale": "LINEAR"
            }
          }
        }
      },
      {
        "width": 6,
        "height": 4,
        "xPos": 6,
        "widget": {
          "title": "Response Latency",
          "xyChart": {
            "dataSets": [
              {
                "timeSeriesQuery": {
                  "timeSeriesFilter": {
                    "filter": "resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"$SERVICE_NAME\" AND metric.type=\"run.googleapis.com/request_latencies\"",
                    "aggregation": {
                      "alignmentPeriod": "60s",
                      "perSeriesAligner": "ALIGN_DELTA",
                      "crossSeriesReducer": "REDUCE_MEAN"
                    }
                  }
                },
                "plotType": "LINE"
              }
            ],
            "timeshiftDuration": "0s",
            "yAxis": {
              "label": "Latency (ms)",
              "scale": "LINEAR"
            }
          }
        }
      },
      {
        "width": 6,
        "height": 4,
        "yPos": 4,
        "widget": {
          "title": "Active Instances",
          "xyChart": {
            "dataSets": [
              {
                "timeSeriesQuery": {
                  "timeSeriesFilter": {
                    "filter": "resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"$SERVICE_NAME\" AND metric.type=\"run.googleapis.com/container/instance_count\"",
                    "aggregation": {
                      "alignmentPeriod": "60s",
                      "perSeriesAligner": "ALIGN_MEAN",
                      "crossSeriesReducer": "REDUCE_SUM"
                    }
                  }
                },
                "plotType": "LINE"
              }
            ],
            "timeshiftDuration": "0s",
            "yAxis": {
              "label": "Instances",
              "scale": "LINEAR"
            }
          }
        }
      },
      {
        "width": 6,
        "height": 4,
        "xPos": 6,
        "yPos": 4,
        "widget": {
          "title": "Database Connections",
          "xyChart": {
            "dataSets": [
              {
                "timeSeriesQuery": {
                  "timeSeriesFilter": {
                    "filter": "resource.type=\"cloudsql_database\" AND metric.type=\"cloudsql.googleapis.com/database/postgresql/num_backends\"",
                    "aggregation": {
                      "alignmentPeriod": "60s",
                      "perSeriesAligner": "ALIGN_MEAN",
                      "crossSeriesReducer": "REDUCE_MEAN"
                    }
                  }
                },
                "plotType": "LINE"
              }
            ],
            "timeshiftDuration": "0s",
            "yAxis": {
              "label": "Connections",
              "scale": "LINEAR"
            }
          }
        }
      },
      {
        "width": 12,
        "height": 4,
        "yPos": 8,
        "widget": {
          "title": "Error Logs",
          "logsPanel": {
            "resourceNames": [
              "projects/$PROJECT_ID"
            ],
            "filter": "resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"$SERVICE_NAME\" AND severity>=ERROR"
          }
        }
      }
    ]
  }
}
EOF
    
    gcloud monitoring dashboards create --config-from-file=dashboard.json
    rm dashboard.json
    
    log_success "Dashboard created"
}

setup_log_based_metrics() {
    log_info "Setting up log-based metrics..."
    
    # Game-specific metrics
    cat > game_start_metric.json << EOF
{
  "name": "projects/$PROJECT_ID/metrics/poker_games_started",
  "description": "Number of poker games started",
  "filter": "resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"$SERVICE_NAME\" AND textPayload:\"Game started\"",
  "metricDescriptor": {
    "metricKind": "GAUGE",
    "valueType": "INT64",
    "displayName": "Poker Games Started"
  }
}
EOF
    
    gcloud logging metrics create poker_games_started --config-from-file=game_start_metric.json
    rm game_start_metric.json
    
    # Player registration metric
    cat > player_registration_metric.json << EOF
{
  "name": "projects/$PROJECT_ID/metrics/player_registrations",
  "description": "Number of player registrations",
  "filter": "resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"$SERVICE_NAME\" AND textPayload:\"User registered\"",
  "metricDescriptor": {
    "metricKind": "GAUGE",
    "valueType": "INT64",
    "displayName": "Player Registrations"
  }
}
EOF
    
    gcloud logging metrics create player_registrations --config-from-file=player_registration_metric.json
    rm player_registration_metric.json
    
    log_success "Log-based metrics created"
}

cleanup() {
    log_info "Cleaning up temporary files..."
    rm -f notification_channel_name.txt
    log_success "Cleanup completed"
}

# Main execution
main() {
    echo
    log_info "📊 Setting up PrimoPoker Monitoring"
    echo
    
    check_prerequisites
    create_notification_channel
    create_uptime_check
    create_alert_policies
    create_dashboard
    setup_log_based_metrics
    cleanup
    
    echo
    log_success "🎉 Monitoring setup completed successfully!"
    echo
    log_info "Access your monitoring dashboard:"
    log_info "https://console.cloud.google.com/monitoring/dashboards?project=$PROJECT_ID"
    echo
    log_info "View alerts:"
    log_info "https://console.cloud.google.com/monitoring/alerting?project=$PROJECT_ID"
    echo
}

# Run the setup
main "$@"
