# Publishing idle-svc

This guide walks **maintainers** through releasing a new version of `idle-svc`, pushing the container image to a registry.  Nothing here is required for end-users—keep it out of the README.

---

## 1. Version bump

1. Decide the semantic version you're releasing (e.g. `v0.2.0`).  
2. Commit with message `release: v0.2.0`.

```bash
git commit -am "release: v0.2.0"
```

---

## 2. Build & push the multi-arch Docker image

We use GitHub Container Registry (GHCR) in this example; adjust for Docker Hub/ECR as needed.

```bash
TAG=v0.2.0
IMAGE="ghcr.io/0xabrar/idle-svc:$TAG"

docker buildx build --platform linux/amd64,linux/arm64 \
  -t "$IMAGE" \
  --push .
```

*Prerequisites*
* `docker buildx` enabled (`docker buildx create --use`).
* Personal Access Token or GitHub Actions with `CR_PAT` scope:
  ```bash
  echo "$CR_PAT" | docker login ghcr.io -u <username> --password-stdin
  ```

---

## 4. Tag & release the source repository

Tag the commit and push:
```bash
git tag $TAG
git push origin $TAG
```

If you use **Goreleaser** or a GitHub Actions workflow, configure it to:
1. Build & attach binaries (`idle-svc_$(GOOS)_$(GOARCH)`).
2. Push the Docker image (step 2). 

A minimal `.github/workflows/release.yml` skeleton is included below; tailor as you wish.

```yaml
name: release
on:
  push:
    tags: [ 'v*' ]

jobs:
  build-image:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.repository_owner }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - run: |
          TAG=${GITHUB_REF##*/}
          docker buildx build --platform linux/amd64,linux/arm64 \
            -t ghcr.io/0xabrar/idle-svc:$TAG \
            -t ghcr.io/0xabrar/idle-svc:latest \
            --push .
```

### Manual image push (no CI)

```bash
# log in once per shell
echo "$CR_PAT" | docker login ghcr.io -u 0xabrar --password-stdin

# push specific version
docker buildx build --platform linux/amd64,linux/arm64 \
  -t ghcr.io/0xabrar/idle-svc:v0.2.0 \
  --push .

# (optional) update latest tag
docker buildx build --platform linux/amd64,linux/arm64 \
  -t ghcr.io/0xabrar/idle-svc:latest \
  --push .
```

### Debugging & local testing

Validate the freshly pushed image against a local kind cluster.

```bash
# create cluster if needed
kind create cluster --name idle-svc-debug
kind get clusters
kubectl cluster-info
kubectl get nodes

# run the image against kind
docker run --rm \
  --network host \
  --user root \
  -v ~/.kube/config:/root/.kube/config:ro \
  ghcr.io/0xabrar/idle-svc:v0.2.0 \
  -A --json
```

The distroless image has no shell; to inspect it:

```bash
# list files
docker run --rm --entrypoint ls ghcr.io/0xabrar/idle-svc:v0.2.0 -R /

# extract binary
docker create --name dump ghcr.io/0xabrar/idle-svc:v0.2.0
docker cp dump:/usr/bin/idle-svc ./idle-svc-debug
docker rm dump && chmod +x idle-svc-debug
```

---

## 5. Post-release tasks

* Bump README examples if they include hard-coded tags
