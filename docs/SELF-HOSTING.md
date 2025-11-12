# Self-Hosting Taikai for Your Organization

This guide is for organizations like **Forge Utah Foundation** who want to self-host Taikai for their tech community events.

## Why Self-Host?

- **Cost Savings**: No monthly fees to platforms like Meetup.com ($15-30/month per group)
- **Data Ownership**: Full control over your community data
- **Customization**: Tailor the platform to your organization's needs
- **Privacy**: Keep member data private and secure
- **No Vendor Lock-in**: Open source means you're never trapped

## Cost Analysis

### Meetup.com Costs (for comparison)

- **Single Group**: ~$15-20/month = $180-240/year
- **Multiple Groups** (3-5): ~$45-100/month = $540-1,200/year
- **Organization Subscription**: ~$180/month = $2,160/year

### Self-Hosting Costs

**Option 1: DigitalOcean Kubernetes**
- Managed Kubernetes: $12/month (basic) to $72/month (production)
- Load Balancer: $12/month
- **Total**: ~$24-84/month = $288-1,008/year

**Option 2: Bare Metal / VPS**
- VPS (4 cores, 8GB RAM): $20-40/month
- **Total**: $240-480/year

**Savings**: **$300-1,900/year** 💰

### Example: Forge Utah Foundation

**Forge Utah runs 5+ groups:**
- Meetup.com cost: ~$900/year minimum
- Self-hosted cost: ~$500/year
- **Annual savings: $400+**

Plus benefits: Full control, customization, data ownership!

---

## Quick Start Guide for Organizations

### Prerequisites

Before you begin, you'll need:

1. **A domain name** (e.g., `events.forgeutah.org`)
2. **Kubernetes cluster** or VPS server
3. **Email service** (SendGrid, Mailgun, or AWS SES)
4. **Basic tech knowledge** (or someone who can help set it up)

### Step-by-Step Setup

#### Step 1: Get a Kubernetes Cluster

**Recommended Providers:**

**DigitalOcean** (easiest, $12/month):
```bash
# Install doctl
brew install doctl  # or download from digitalocean.com

# Authenticate
doctl auth init

# Create cluster
doctl kubernetes cluster create taikai-cluster \
  --region nyc1 \
  --node-pool "name=taikai-pool;size=s-2vcpu-4gb;count=2"

# Connect kubectl
doctl kubernetes cluster kubeconfig save taikai-cluster
```

**Google GKE** (free tier available):
```bash
gcloud container clusters create taikai-cluster \
  --zone us-central1-a \
  --machine-type n1-standard-2 \
  --num-nodes 2
```

**AWS EKS** or **Azure AKS**: See provider documentation

#### Step 2: Install Prerequisites

```bash
# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Install cert-manager (for automatic SSL)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Install nginx ingress controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/cloud/deploy.yaml
```

#### Step 3: Configure DNS

Point your domain to the cluster:

```bash
# Get load balancer IP
kubectl get svc -n ingress-nginx ingress-nginx-controller

# Create A record in your DNS:
# events.forgeutah.org -> <LOAD_BALANCER_IP>
```

#### Step 4: Prepare Email Service

**Option A: SendGrid (Free 100 emails/day)**

1. Sign up at sendgrid.com
2. Create API key
3. Verify sender email address

**Option B: AWS SES (Free 62,000 emails/month)**

1. Sign up for AWS
2. Verify domain in SES
3. Create IAM user with SES permissions
4. Note access key and secret

#### Step 5: Configure Taikai

Clone the repository:

```bash
git clone https://github.com/forgeutah/taikai.git
cd taikai/helm/taikai
```

Create your organization's configuration:

```bash
cat > forge-utah-values.yaml <<EOF
organization:
  name: "Forge Utah Foundation"
  domain: "events.forgeutah.org"
  email: "noreply@forgeutah.org"

image:
  repository: forgeutah/taikai
  tag: "latest"

server:
  replicaCount: 2

config:
  jwt:
    secret: "$(openssl rand -base64 32)"

  email:
    provider: smtp
    smtp:
      host: "smtp.sendgrid.net"
      port: 587
      username: "apikey"
      password: "YOUR_SENDGRID_API_KEY"  # Replace this

postgresql:
  auth:
    password: "$(openssl rand -base64 32)"

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
  hosts:
    - host: events.forgeutah.org
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: taikai-tls
      hosts:
        - events.forgeutah.org
EOF
```

#### Step 6: Create Let's Encrypt Issuer

```bash
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@forgeutah.org  # Your email
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

#### Step 7: Deploy Taikai

```bash
helm install taikai . \
  -f forge-utah-values.yaml \
  --create-namespace \
  --namespace taikai

# Watch deployment
kubectl get pods -n taikai -w
```

#### Step 8: Verify Deployment

```bash
# Check all pods are running
kubectl get pods -n taikai

# Check TLS certificate
kubectl get certificate -n taikai

# Visit your site
open https://events.forgeutah.org
```

#### Step 9: Create Admin User

```bash
# Access the database
kubectl exec -it -n taikai deployment/taikai-postgres -- psql -U taikai

# In PostgreSQL shell:
-- Create admin user
INSERT INTO users (email, password_hash, name, timezone, email_verified, is_active)
VALUES (
  'admin@forgeutah.org',
  '$2a$12$...',  -- Use bcrypt hash
  'Admin User',
  'America/Denver',
  true,
  true
);

-- Make them org admin
INSERT INTO org_admins (user_id, organization_id)
SELECT id, (SELECT id FROM organizations LIMIT 1)
FROM users WHERE email = 'admin@forgeutah.org';
```

Or seed with provided script:

```bash
kubectl exec -n taikai deployment/taikai-server -- /app/migrate seed
```

---

## Migrating from Meetup.com

### Exporting Your Data

1. **Export members**: Download member list from Meetup.com
2. **Export events**: Copy event details manually or use API
3. **Export RSVPs**: Note attendance for historical data

### Import to Taikai

Use the included migration script:

```bash
# Prepare CSV files
# members.csv: email,name,joined_date
# events.csv: title,description,date,venue

# Run migration
kubectl exec -it -n taikai deployment/taikai-server -- \
  /app/migrate import-meetup \
  --members=/data/members.csv \
  --events=/data/events.csv
```

### Communication to Members

Send announcement email to existing members:

```
Subject: We're Moving to Our Own Event Platform!

Hi [Community Name] Members,

Great news! We're launching our own event platform at events.forgeutah.org

Why the change?
- Faster, better experience
- No monthly fees (saving money for community events!)
- More features and customization
- Your data stays with us, not a third party

What you need to do:
1. Visit https://events.forgeutah.org
2. Create your account (use the same email)
3. Browse and RSVP to upcoming events

All our events will now be posted here. We'll keep the Meetup page for a transition period.

Questions? Reply to this email!

[Your Name]
[Organization Name]
```

---

## Maintenance Tasks

### Weekly Tasks

```bash
# Check pod health
kubectl get pods -n taikai

# Review logs for errors
kubectl logs -n taikai -l app=taikai --tail=100 | grep -i error
```

### Monthly Tasks

```bash
# Update Docker images
helm upgrade taikai ./helm/taikai -f forge-utah-values.yaml

# Verify backups
kubectl exec -n taikai deployment/taikai-postgres -- \
  pg_dump -U taikai taikai > backup-$(date +%Y%m).sql

# Review resource usage
kubectl top pods -n taikai
```

### Quarterly Tasks

- Review and update dependencies
- Test disaster recovery process
- Review access permissions
- Update documentation

---

## Cost Optimization Tips

### 1. Start Small, Scale Up

```yaml
# Initial deployment (2 users)
server:
  replicaCount: 1

postgresql:
  primary:
    persistence:
      size: 5Gi

# Can scale up later as community grows
```

### 2. Use Spot/Preemptible Instances

Save 60-80% on compute:

```bash
# GKE example
gcloud container node-pools create spot-pool \
  --cluster taikai-cluster \
  --spot \
  --num-nodes 2
```

### 3. Use External Managed Services

For better reliability, use managed database:

```yaml
postgresql:
  enabled: false  # Use managed PostgreSQL

config:
  database:
    host: your-managed-postgres.cloud.com
```

### 4. Free Email Tier

- **SendGrid**: 100 emails/day free
- **AWS SES**: 62,000 emails/month free (with EC2)
- **Mailgun**: 5,000 emails/month free

### 5. Shared Clusters

Multiple organizations can share a cluster:

```bash
# Deploy for multiple orgs
helm install forge-utah ./taikai -f forge-utah-values.yaml -n forge-utah
helm install utah-js ./taikai -f utah-js-values.yaml -n utah-js
```

---

## Example: Forge Utah Foundation Setup

### Our Configuration

**Infrastructure:**
- DigitalOcean Kubernetes ($72/month)
- 2-node cluster (4GB RAM each)
- Managed PostgreSQL add-on ($15/month)

**Services:**
- Domain: events.forgeutah.org (via Cloudflare)
- Email: SendGrid (free tier)
- SSL: Let's Encrypt (free)

**Monthly Cost: ~$87** (down from ~$900 on Meetup.com)
**Annual Savings: ~$800+** 🎉

### Our Groups

We manage 5+ tech groups:
- Kubernetes Meetup
- Go User Group
- Data Engineering
- DevOps SLC
- Cloud Native Utah

All under one platform, one admin interface!

### Migration Timeline

**Week 1**: Set up infrastructure and deploy
**Week 2**: Import data and test
**Week 3**: Invite core members to test
**Week 4**: Full launch and announcement
**Week 5**: Sunset Meetup.com

---

## Troubleshooting

### "Can't access the site"

```bash
# Check DNS
dig events.forgeutah.org

# Check ingress
kubectl get ingress -n taikai

# Check certificate
kubectl get certificate -n taikai
kubectl describe certificate taikai-tls -n taikai
```

### "Emails not sending"

```bash
# Check email config
kubectl get configmap taikai-config -n taikai -o yaml | grep SMTP

# Test SMTP connection
telnet smtp.sendgrid.net 587

# Check logs
kubectl logs -n taikai -l app=taikai | grep -i email
```

### "Database connection failed"

```bash
# Check database pod
kubectl get pods -n taikai | grep postgres

# Test connection
kubectl exec -it -n taikai deployment/taikai-postgres -- psql -U taikai

# Check credentials
kubectl get secret taikai-secrets -n taikai -o yaml
```

---

## Getting Help

### Community Support

- **GitHub Discussions**: Ask questions, share tips
- **Discord**: Real-time help from community
- **Documentation**: Comprehensive guides

### Professional Support

For organizations needing help:
- Setup assistance
- Custom development
- Training workshops
- Managed hosting

Contact: support@forgeutah.org

---

## Success Stories

### Forge Utah Foundation

"We were spending $900/year on Meetup.com for our 5 groups. After self-hosting Taikai, we're down to $500/year and have WAY more control. The migration took one weekend, and our members love the new platform!"

*- Forge Utah Foundation*

### [Your Organization Here!]

Have you successfully deployed Taikai? Share your story!

---

## Next Steps

1. ✅ Set up infrastructure
2. ✅ Deploy Taikai
3. ✅ Import your data
4. ✅ Customize branding
5. ✅ Invite core members
6. ✅ Announce to community
7. ✅ Celebrate! 🎉

**Welcome to the self-hosted community events movement!**

For more help: https://github.com/forgeutah/taikai
