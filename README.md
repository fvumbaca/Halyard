# Halyard

> Halyard is in an experimentation phase where I will will be learning a lot
and changing my opinion (and APIs) very often.

Halyard is a deployment and resource management tool. It packs an opinionated
collection of best-in-class features of other similar tools with an emphasis on
simplicity and efficiency.

## Install

```sh
go install github.com/fvumbaca/halyard@latest
```

## Usage

Base Deployment (`base.yaml`):
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx
spec:
  selector:
    app: nginx
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
```

Production Patches (`prod.yaml`):
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
  annotations:
    halyard.sh/layer: "prod"
spec:
  replicas: 3
  template:
    spec:
      containers:
      - image: nginx:1.14.2
```

Preview rendered resources:
```shell
halyard template base.yaml prod.yaml
```

Update/Create resources:
```shell
halyard apply base.yaml prod.yaml
```

## TODO

- [ ] Better cli flags/options
  - [ ] Default namespace override
- [ ] Layer filtering
- [ ] Resource association via selection labels
  - [ ] Resource cleanup
- [ ] Value based templating
  - [ ] Use variables/patching
- [ ] Apply Methods
  - [ ] Local and server-side resource validation/dry running
  - [ ] Look into server side apply
  - [ ] Look into patching over updating
- [ ] Track update history
  - [ ] store changes to resources in a history file or configmap
