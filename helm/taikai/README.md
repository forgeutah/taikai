# Taikai Helm Chart

This Helm chart deploys Taikai, an open-source event management platform for tech communities.

## Prerequisites

- Kubernetes 1.20+
- Helm 3.0+
- PV provisioner support in the underlying infrastructure (for PostgreSQL and Redis persistence)
- Ingress controller (nginx recommended)
- cert-manager (for automatic TLS certificates)

## Installing the Chart

### Quick Start

```bash
# Add the Taikai Helm repository (when available)
# helm repo add taikai https://forgeutah.github.io/taikai-helm

# For now, install from local chart
cd helm/taikai

# Create a custom values file
cat > my-values.yaml <<EOF
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

# Install the chart
helm install taikai . -f my-values.yaml --create-namespace --namespace taikai
```

## Configuration

The following table lists the configurable parameters and their default values.

### Organization Settings

| Parameter | Description | Default |
|-----------|-------------|---------|
| `organization.name` | Organization name | `Forge Utah Foundation` |
| `organization.domain` | Domain for the platform | `events.forgeutah.org` |
| `organization.email` | Email address for system emails | `noreply@forgeutah.org` |

### Application Settings

| Parameter | Description | Default |
|-----------|-------------|---------|
| `server.replicaCount` | Number of server replicas | `2` |
| `worker.replicaCount` | Number of worker replicas | `1` |
| `config.jwt.secret` | JWT secret key (MUST change) | `CHANGEME...` |
| `config.jwt.accessTokenExpiry` | Access token expiry in seconds | `900` |
| `config.jwt.refreshTokenExpiry` | Refresh token expiry in seconds | `604800` |

### Database Settings

| Parameter | Description | Default |
|-----------|-------------|---------|
| `postgresql.enabled` | Enable built-in PostgreSQL | `true` |
| `postgresql.auth.username` | PostgreSQL username | `taikai` |
| `postgresql.auth.password` | PostgreSQL password | `CHANGEME...` |
| `postgresql.auth.database` | PostgreSQL database name | `taikai` |
| `postgresql.primary.persistence.size` | PostgreSQL storage size | `10Gi` |

### Redis Settings

| Parameter | Description | Default |
|-----------|-------------|---------|
| `redis.enabled` | Enable built-in Redis | `true` |
| `redis.master.persistence.size` | Redis storage size | `5Gi` |

### Ingress Settings

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `true` |
| `ingress.className` | Ingress class name | `nginx` |
| `ingress.hosts[0].host` | Hostname | `events.forgeutah.org` |
| `ingress.tls[0].secretName` | TLS secret name | `taikai-tls` |

## Upgrading

```bash
# Update values
vim my-values.yaml

# Upgrade the release
helm upgrade taikai . -f my-values.yaml --namespace taikai
```

## Uninstalling

```bash
helm uninstall taikai --namespace taikai
kubectl delete namespace taikai
```

## Using External Database

To use an external PostgreSQL database:

```yaml
postgresql:
  enabled: false

# Add to ConfigMap
config:
  database:
    host: my-postgres-host.example.com
    port: 5432
    name: taikai
    user: taikai
    password: my-secure-password
```

## Using External Redis

To use an external Redis:

```yaml
redis:
  enabled: false

# Add to ConfigMap
config:
  redis:
    host: my-redis-host.example.com
    port: 6379
```

## Email Configuration

### Using SMTP (SendGrid, Mailgun, etc.)

```yaml
config:
  email:
    provider: smtp
    smtp:
      host: smtp.sendgrid.net
      port: 587
      username: apikey
      password: your-sendgrid-api-key
```

### Using AWS SES

```yaml
config:
  email:
    provider: ses
    ses:
      region: us-east-1
      accessKey: your-access-key
      secretKey: your-secret-key
```

## Monitoring

To enable Prometheus monitoring:

```yaml
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
```

## Troubleshooting

### Check pod status

```bash
kubectl get pods -n taikai
```

### View logs

```bash
# Server logs
kubectl logs -n taikai -l app=taikai,component=server -f

# Worker logs
kubectl logs -n taikai -l app=taikai,component=worker -f

# Database logs
kubectl logs -n taikai -l app=postgres -f
```

### Check configuration

```bash
kubectl get configmap taikai-config -n taikai -o yaml
kubectl get secret taikai-secrets -n taikai -o yaml
```

## Support

For issues and questions:
- GitHub Issues: https://github.com/forgeutah/taikai/issues
- Documentation: https://github.com/forgeutah/taikai/tree/main/docs
