# Taikai Deployment Guide

This guide covers deploying Taikai to production using Docker and Kubernetes.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Deployment Options](#deployment-options)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Configuration](#configuration)
- [SSL/TLS Setup](#ssltls-setup)
- [Database Management](#database-management)
- [Monitoring](#monitoring)
- [Backup and Recovery](#backup-and-recovery)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Infrastructure Requirements

**Minimum Resources:**
- 2 CPU cores
- 4 GB RAM
- 20 GB storage

**Recommended Resources (for ~500 users):**
- 4 CPU cores
- 8 GB RAM
- 50 GB storage

### Software Requirements

- Docker 20.10+ (for Docker deployment)
- Kubernetes 1.20+ (for K8s deployment)
- Helm 3.0+ (for Helm deployment)
- PostgreSQL 15+ (if using external database)
- Redis 7+ (if using external cache)

### External Services (Optional)

- **Email Service**: AWS SES, SendGrid, Mailgun, or any SMTP provider
- **SMS Service**: Twilio (for Phase 2)
- **DNS**: For custom domain
- **TLS Certificates**: Let's Encrypt (free) or commercial certificate

---

## Deployment Options

Taikai can be deployed in several ways:

1. **Docker Compose** - Simple, single-server deployment
2. **Kubernetes (raw manifests)** - Full control, manual setup
3. **Helm Chart** - Kubernetes deployment, simplified configuration
4. **Managed Kubernetes** - GKE, EKS, AKS, DigitalOcean K8s

We recommend **Helm** for most organizations.

---

## Docker Deployment

### Quick Start with Docker Compose

**For production use, see the Kubernetes section below.**

1. **Clone the repository:**

```bash
git clone https://github.com/forgeutah/taikai.git
cd taikai
```

2. **Configure environment:**

```bash
cp .env.example .env
nano .env  # Edit configuration
```

Update these critical values:
- `DATABASE_PASSWORD` - Set a strong password
- `JWT_SECRET` - Use `openssl rand -base64 32`
- Email settings (SMTP or AWS SES)

3. **Build and run:**

```bash
docker-compose up -d
```

4. **Run migrations:**

```bash
docker-compose exec server /app/migrate
```

5. **Access the application:**

Open http://localhost:8080

### Production Docker Deployment

**Not recommended for production.** Use Kubernetes instead for:
- High availability
- Auto-scaling
- Rolling updates
- Better resource management

---

## Kubernetes Deployment

### Option 1: Using Helm (Recommended)

#### Step 1: Install Prerequisites

```bash
# Install cert-manager for automatic TLS
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create ClusterIssuer for Let's Encrypt
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@yourdomain.com  # Change this
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

#### Step 2: Create Values File

```bash
cd helm/taikai

cat > my-org-values.yaml <<EOF
organization:
  name: "My Organization"
  domain: "events.myorg.com"
  email: "noreply@myorg.com"

config:
  jwt:
    secret: "$(openssl rand -base64 32)"

postgresql:
  auth:
    password: "$(openssl rand -base64 32)"

config:
  email:
    provider: smtp
    smtp:
      host: "smtp.sendgrid.net"
      port: 587
      username: "apikey"
      password: "YOUR_SENDGRID_API_KEY"

ingress:
  hosts:
    - host: events.myorg.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: taikai-tls
      hosts:
        - events.myorg.com
EOF
```

#### Step 3: Install the Chart

```bash
helm install taikai . \
  -f my-org-values.yaml \
  --create-namespace \
  --namespace taikai
```

#### Step 4: Verify Deployment

```bash
# Check pod status
kubectl get pods -n taikai

# Check services
kubectl get svc -n taikai

# Check ingress
kubectl get ingress -n taikai

# View logs
kubectl logs -n taikai -l app=taikai,component=server -f
```

### Option 2: Using Raw Kubernetes Manifests

#### Step 1: Update Configuration

```bash
cd k8s/base

# Edit configmap.yaml
nano configmap.yaml
# Update: BASE_URL, SMTP_FROM_EMAIL, etc.

# Edit secret.yaml
nano secret.yaml
# Update: DATABASE_PASSWORD, JWT_SECRET, etc.
```

#### Step 2: Apply Manifests

```bash
kubectl apply -f namespace.yaml
kubectl apply -f secret.yaml
kubectl apply -f configmap.yaml
kubectl apply -f postgres.yaml
kubectl apply -f redis.yaml
kubectl apply -f server.yaml
kubectl apply -f worker.yaml
```

#### Step 3: Wait for Pods

```bash
kubectl wait --for=condition=ready pod -l app=taikai -n taikai --timeout=300s
```

---

## Configuration

### Environment Variables

See `.env.example` for all available options.

**Critical Variables:**

```bash
# Application
APP_ENV=production
BASE_URL=https://events.yourdomain.com

# Security
JWT_SECRET=<random-32-char-string>

# Database
DATABASE_URL=postgresql://user:password@host:5432/taikai

# Email (choose one)
# Option 1: SMTP
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your-api-key

# Option 2: AWS SES
AWS_SES_REGION=us-east-1
AWS_SES_ACCESS_KEY=your-access-key
AWS_SES_SECRET_KEY=your-secret-key
```

### Generating Secrets

```bash
# JWT Secret
openssl rand -base64 32

# Database Password
openssl rand -base64 32

# Or use pwgen
pwgen -s 32 1
```

---

## SSL/TLS Setup

### Using cert-manager (Recommended)

The Helm chart includes cert-manager integration:

```yaml
ingress:
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  tls:
    - secretName: taikai-tls
      hosts:
        - events.yourdomain.com
```

Certificates are automatically provisioned and renewed.

### Using Custom Certificates

```bash
# Create TLS secret
kubectl create secret tls taikai-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key \
  --namespace=taikai
```

---

## Database Management

### Running Migrations

Migrations run automatically via init container in Kubernetes.

**Manual migration:**

```bash
# Using kubectl
kubectl exec -it -n taikai deployment/taikai-server -- /app/migrate

# Using Helm
helm upgrade taikai . -f my-values.yaml --namespace taikai
```

### Database Backups

#### Automated Backups (Recommended)

```bash
# Create CronJob for daily backups
cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
  namespace: taikai
spec:
  schedule: "0 2 * * *"  # 2 AM daily
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:15-alpine
            command:
            - /bin/sh
            - -c
            - |
              pg_dump -h taikai-postgres -U taikai taikai | gzip > /backup/taikai-\$(date +%Y%m%d-%H%M%S).sql.gz
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: taikai-secrets
                  key: DATABASE_PASSWORD
            volumeMounts:
            - name: backup
              mountPath: /backup
          volumes:
          - name: backup
            persistentVolumeClaim:
              claimName: backup-pvc
          restartPolicy: OnFailure
EOF
```

#### Manual Backup

```bash
# Backup
kubectl exec -n taikai deployment/taikai-postgres -- \
  pg_dump -U taikai taikai | gzip > taikai-backup-$(date +%Y%m%d).sql.gz

# Restore
gunzip -c taikai-backup-20250112.sql.gz | \
  kubectl exec -i -n taikai deployment/taikai-postgres -- \
  psql -U taikai taikai
```

---

## Monitoring

### Health Checks

```bash
# Check application health
curl https://events.yourdomain.com/health

# Should return: {"status":"ok"}
```

### Logs

```bash
# Server logs
kubectl logs -n taikai -l app=taikai,component=server -f

# Worker logs
kubectl logs -n taikai -l app=taikai,component=worker -f

# Database logs
kubectl logs -n taikai -l app=postgres -f

# All pods
kubectl logs -n taikai --all-containers=true -f
```

### Prometheus Monitoring (Optional)

Enable in Helm values:

```yaml
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
```

---

## Backup and Recovery

### Disaster Recovery Plan

1. **Regular Backups:**
   - Database: Daily automated backups
   - Uploaded files: Daily sync to S3/B2
   - Configuration: Store in git

2. **Backup Storage:**
   - Keep 7 daily backups
   - Keep 4 weekly backups
   - Keep 3 monthly backups

3. **Recovery Procedure:**

```bash
# 1. Deploy fresh Taikai instance
helm install taikai-recovery ./helm/taikai -f backup-values.yaml

# 2. Restore database
gunzip -c backup.sql.gz | kubectl exec -i ... psql ...

# 3. Restore uploaded files
kubectl cp uploads/ taikai-server:/app/uploads/

# 4. Verify functionality
curl https://events-recovery.yourdomain.com/health
```

---

## Troubleshooting

### Pod Not Starting

```bash
# Check pod status
kubectl describe pod -n taikai <pod-name>

# Check events
kubectl get events -n taikai --sort-by='.lastTimestamp'

# Check logs
kubectl logs -n taikai <pod-name>
```

### Database Connection Issues

```bash
# Test database connectivity
kubectl exec -it -n taikai deployment/taikai-postgres -- psql -U taikai

# Check database credentials
kubectl get secret taikai-secrets -n taikai -o yaml
```

### Email Not Sending

```bash
# Check email configuration
kubectl get configmap taikai-config -n taikai -o yaml

# Check server logs for email errors
kubectl logs -n taikai -l app=taikai,component=server | grep -i email

# Test SMTP connection
kubectl run -it --rm debug --image=alpine --restart=Never -- \
  sh -c "apk add curl && curl -v telnet://smtp.sendgrid.net:587"
```

### Performance Issues

```bash
# Check resource usage
kubectl top pods -n taikai

# Scale server replicas
kubectl scale deployment taikai-server --replicas=4 -n taikai

# Or with Helm
helm upgrade taikai . --set server.replicaCount=4 -f my-values.yaml
```

### Ingress Not Working

```bash
# Check ingress status
kubectl get ingress -n taikai
kubectl describe ingress taikai-ingress -n taikai

# Check cert-manager certificates
kubectl get certificate -n taikai
kubectl describe certificate taikai-tls -n taikai

# Check DNS
dig events.yourdomain.com
```

---

## Security Checklist

Before going to production:

- [ ] Change all default passwords
- [ ] Generate new JWT secret
- [ ] Enable HTTPS/TLS
- [ ] Configure firewall rules
- [ ] Set up database backups
- [ ] Enable pod security policies
- [ ] Restrict network access
- [ ] Use secrets management (Sealed Secrets, External Secrets)
- [ ] Enable audit logging
- [ ] Set resource limits
- [ ] Configure RBAC properly
- [ ] Regular security updates

---

## Scaling

### Horizontal Scaling

```bash
# Scale server replicas
kubectl scale deployment taikai-server --replicas=5 -n taikai

# Enable autoscaling
kubectl autoscale deployment taikai-server \
  --min=2 --max=10 \
  --cpu-percent=80 \
  -n taikai
```

### Vertical Scaling

Update resources in values.yaml:

```yaml
server:
  resources:
    requests:
      memory: "512Mi"
      cpu: "500m"
    limits:
      memory: "1Gi"
      cpu: "2000m"
```

---

## Support

For deployment help:
- GitHub Issues: https://github.com/forgeutah/taikai/issues
- Documentation: https://github.com/forgeutah/taikai/tree/main/docs
- Community: https://discord.gg/forgeutah
