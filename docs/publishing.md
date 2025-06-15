# Publishing idle-svc

This guide walks **maintainers** through releasing a new version of `idle-svc`, pushing the container image to a registry, and distributing the Helm chart.  Nothing here is required for end-users—keep it out of the README.

---

## 1. Version bump

1. Decide the semantic version you're releasing (e.g. `v0.2.0`).  
2. Update `chart/idle-svc/Chart.yaml` `version` **and** `appVersion`.  
3. Commit with message `release: v0.2.0`.

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

## 3. Package the Helm chart

The chart is located at `chart/idle-svc/`.

```bash
helm lint chart/idle-svc
helm package chart/idle-svc -d dist
```

This produces `dist/idle-svc-$TAG.tgz`.

### Option A – GitHub Pages index (classic repo)

```bash
# clone/pull the gh-pages branch into ./charts-site (or another dir)
mkdir -p charts-site && cp dist/idle-svc-$TAG.tgz charts-site/
cd charts-site
helm repo index . --url https://0xabrar.github.io/idle-svc-charts
# commit & push to gh-pages
```

End-users then:
```bash
helm repo add idle-svc https://0xabrar.github.io/idle-svc-charts
helm install idle-svc idle-svc/idle-svc --version $TAG
```

### Option B – OCI registry (Helm 3.8+)

```bash
helm push dist/idle-svc-$TAG.tgz oci://ghcr.io/0xabrar/charts
```

End-users then:
```bash
helm pull oci://ghcr.io/0xabrar/charts/idle-svc --version $TAG
helm install idle-svc idle-svc-$TAG.tgz
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
3. Push the Helm chart (step 3).

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

  helm-chart:
    runs-on: ubuntu-latest
    needs: build-image
    steps:
      - uses: actions/checkout@v4
      - run: |
          TAG=${GITHUB_REF##*/}
          helm package chart/idle-svc -d dist --version $TAG --app-version $TAG
          helm push dist/idle-svc-$TAG.tgz oci://ghcr.io/0xabrar/charts
```

---

## 5. Post-release tasks

* Update `values.yaml` default `image.tag` to the new version for the next development cycle.  
* Open a PR to bump README examples if they include hard-coded tags.
* If using gh-pages, confirm the updated `index.yaml` is accessible.

---

That's it—your image and chart are now publicly installable. 🎉 