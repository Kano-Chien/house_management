# Kubernetes deployment

This directory contains a minimal single-node friendly Kubernetes setup for House Management.

The app is packaged as one container that serves both:

- the Vue frontend from `frontend/dist`
- the Go backend/API on port `8080`

SQLite and uploaded images are persisted through a `PersistentVolumeClaim` mounted at `/app/data` and `/app/uploads`.

## 1. Build the image locally

For kind/minikube/k3d testing, build an image first:

```bash
docker build -t ghcr.io/kano-chien/house-management:latest .
```

If you use kind, load the image into the cluster:

```bash
kind load docker-image ghcr.io/kano-chien/house-management:latest
```

If you use minikube, build inside minikube's Docker daemon instead:

```bash
eval $(minikube docker-env)
docker build -t ghcr.io/kano-chien/house-management:latest .
```

## 2. Configure secrets

`secret.example.yaml` creates an empty `LINE_NOTIFY_TOKEN` so the app can start without LINE Notify.

For a real token, edit the file before applying or create your own secret:

```bash
kubectl create secret generic house-management-secret \
  -n house-management \
  --from-literal=LINE_NOTIFY_TOKEN='your-token'
```

## 3. Deploy

```bash
kubectl apply -k k8s
```

Check status:

```bash
kubectl get pods,svc,ingress,pvc -n house-management
kubectl logs -n house-management deploy/house-management
```

## 4. Access the app

The included Ingress uses:

```text
house-management.local
```

For local testing, either configure your ingress controller and `/etc/hosts`, or use port-forwarding:

```bash
kubectl port-forward -n house-management svc/house-management 8080:80
```

Then open:

```text
http://localhost:8080
```

## 5. Update the image

After building and publishing a new image tag, update `k8s/kustomization.yaml`:

```yaml
images:
  - name: ghcr.io/kano-chien/house-management
    newTag: your-new-tag
```

Then apply again:

```bash
kubectl apply -k k8s
kubectl rollout status -n house-management deployment/house-management
```

Rollback if needed:

```bash
kubectl rollout undo -n house-management deployment/house-management
```

## Notes

- This app uses SQLite, so the Deployment intentionally runs with `replicas: 1`.
- Do not scale this Deployment above 1 unless the database layer is changed away from a single SQLite file.
- The PVC stores `/app/data/house.db` and `/app/uploads` so data survives Pod restarts.
- The current probes use `/` because the Go server serves the built SPA there.
