# Deploy a [DigitalOcean App Platform](https://www.digitalocean.com/products/app-platform/) app using GitHub Actions.

Deploy an app from source (including the configuration) on commit, while allowing you to run tests or perform other operations as part of your CI/CD pipeline.

- Supports picking up an in-repository (or filesystem really) `app.yaml` (defaults to `.do/app.yaml`, configurable via the `app_spec_location` input) to create the app from instead of having to rely on an already existing app that's then downloaded (though that is still supported). The in-filesystem app spec can also be templated with environment variables automatically (see examples below).
- Prints the build and deploy logs into the Github Action log on demand (configurable via `print_build_logs` and `print_deploy_logs`) and surfaces them as outputs `build_logs` and `deploy_logs`.
- Provides the app's metadata as the output `app`.
- Supports a "preview mode" geared towards orchestrating per-PR app previews. It can be enabled via `deploy_pr_review`, see the [Implementing Preview Apps](#implementing-preview-apps) example.

## Support

If you require assistance or have a feature idea, please create a support ticket at the [official DigitalOcean Support](https://cloudsupport.digitalocean.com/s/).

## Documentation

### `deploy` action

#### Inputs

- `token`: DigitalOcean Personal Access Token. See https://docs.digitalocean.com/reference/api/create-personal-access-token/ for creating a new token.
- `app_spec_location`: Location of the app spec file. Defaults to `.do/app.yaml`.
- `app_name`: Name of the app to pull the spec from. The app must already exist. If an app name is given, a potential in-repository app spec is ignored.
- `print_build_logs`: Print build logs. Defaults to `false`.
- `print_deploy_logs`: Print deploy logs. Defaults to `false`.
- `deploy_pr_preview`: Deploy the app as a PR preview. The app name will be derived from the PR, the app spec will be modified to exclude conflicting configuration like domains and alerts and all Github references to the current repository will be updated to point to the PR's branch. Defaults to `false`.

#### Outputs

- `app`: A JSON representation of the entire app after the deployment.
- `build_logs`: The builds logs of the deployment.
- `deploy_logs`: The deploy logs of the deployment.

### `delete` action

#### Inputs

- `token`: DigitalOcean Personal Access Token. See https://docs.digitalocean.com/reference/api/create-personal-access-token/ for creating a new token.
- `app_id`: ID of the app to delete.
- `app_name`: Name of the app to delete.
- `from_pr_preview`: Use this if the app was deployed as a PR preview. The app name will be derived from the PR and.
- `ignore_not_found`: Ignore if the app is not found.

## Usage

As a prerequisite for all examples, you'll need a `DIGITALOCEAN_ACCESS_TOKEN`[secret](https://docs.github.com/en/actions/reference/encrypted-secrets#creating-encrypted-secrets-for-a-repository) in the respective repository. If not already done, get a DigitalOcean Personal Access token by following this [instructions](https://docs.digitalocean.com/reference/api/create-personal-access-token/) and declare it as that secret in the repository you're working with.

### Deploy an app

With the following contents of `.do/app.yaml` in the repository:

```yaml
name: sample
services:
- name: sample
  github:
    branch: main
    repo: digitalocean/app_action_example
```

The following action deploys the app whenever a new commit is pushed to the main branch. Note that `deploy_on_push` is **not** used here, since the Github Action is the driving force behind the deployment. Also note that updates to `.do/app.yaml` will automatically be applied to the app.

```yaml
name: Update App

on:
  push:
    branches: [main]

jobs:
  deploy-app:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
      - name: Deploy the app
        uses: digitalocean/app_action/deploy@main
```

### Deploy an app with a prebuilt image

With the following contents of `.do/app.yaml` in the repository:

```yaml
name: sample
services:
- name: sample
  image:
    registry_type: GHCR
    registry: digitalocean
    repository: app_action_example
    digest: ${SAMPLE_DIGEST}
```

The following action builds a new image from a Dockerfile in the repository and deploys the respective app from it. The build in App Platform is automatically bypassed. The built image is deployed from its digest, avoiding any inconsistencies around mutable tags and guaranteeing that **exactly** this image is deployed.

```yaml
name: Build, Push and Deploy a Docker Image

on:
  push:
    branches: [main]

jobs:
  build-push-deploy-image:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
      id-token: write
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
      - name: Log in to the Container registry
        uses: docker/login-action@65b78e6e13532edd9afa3aa52ac7964289d1a9c1
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: Build and push Docker image
        id: push
        uses: docker/build-push-action@f2a1d5e99d037542a71f64918e516c093c6f3fc4
        with:
          context: .
          push: true
          tags: ghcr.io/${{ github.repository }}:latest
      - name: Deploy the app
        uses: digitalocean/app_action/deploy@main
        env:
          SAMPLE_DIGEST: ${{ steps.push.outputs.digest }}
        with:
          token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
```

### Launch a preview app per pull request

With the following 2 actions, you can implement a "Preview Apps" feature, that provide a per-PR app to check if the deployment **would** work. If the deployment succeeds, a comment is posted with the live URL of the app. If the deployment fails, a link to the respective action run is posted alongside the build and deployment logs for quick debugging.

Once the PR is closed or merged, the respective app is deleted again.

```yaml
name: App Platform Preview

on:
  pull_request:
    branches: [main]

permissions:
  pull-requests: write

jobs:
  test:
    name: preview
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
      - name: Deploy the app
        id: deploy
        uses: digitalocean/app_action/deploy@main
        with:
          deploy_pr_preview: "true"
          token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
      - uses: actions/github-script@v7
        env:
          BUILD_LOGS: ${{ steps.deploy.outputs.build_logs }}
          DEPLOY_LOGS: ${{ steps.deploy.outputs.deploy_logs }}
        with:
          script: |
            const { BUILD_LOGS, DEPLOY_LOGS } = process.env
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `:rocket: :rocket: :rocket: The app was successfully deployed at ${{ fromJson(steps.deploy.outputs.app).live_url }}.`
            })
      - uses: actions/github-script@v7
        if: failure()
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `The app failed to be deployed. Logs can be found [here](https://github.com/${{ github.repository }}/actions/runs/${{ github.run_id }}).

              ## Logs
              <details>
              <summary>Build logs</summary>

              \`\`\`
              ${BUILD_LOGS}
              \`\`\`
              </details>

              <details>
              <summary>Deploy logs</summary>

              \`\`\`
              ${DEPLOY_LOGS}
              \`\`\`
              </details>`
            })
```

```yaml
name: Delete Preview

on:
  pull_request:
    types: [ closed ]

jobs:
  closed:
    runs-on: ubuntu-latest
    steps:
      - name: delete preview app
        uses: digitalocean/app_action/delete@main
        with:
          from_pr_preview: "true"
          ignore_not_found: "true"
          token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
```

## Note for handling container images

It is strongly suggested to use image digests to identify a specific image like in the example above. If that is not possible, it is strongly suggested to use a unique and descriptive tag for the respective image (not `latest`).

## Upgrade from v1.x

The v1 branch of this action is no longer under active development. Its documentation is [still available](https://github.com/digitalocean/app_action/blob/v1/README.md) though.

The new deploy action does not support the `images` input from the old action. For in-repository app specs, it's suggested to use env-var-substitution as in the example above. If the spec of an existing app should be updated via the backwards-compatible `app_name` input, the `IMAGE_DIGEST_$component-name`/`IMAGE_TAG_$component-name` environment variables can be used to change the respective fields of the image reference.

## Resources to know more about DigitalOcean App Platform App Spec

- [App Platform Guided App Spec Declaration](https://www.digitalocean.com/community/tech_talks/defining-your-app-specification-on-digitalocean-app-platform)
- [App Platform App Spec Blog](https://docs.digitalocean.com/products/app-platform/reference/app-spec/)
- [App Platform App Spec Components](https://www.digitalocean.com/blog/build-component-based-apps-with-digitalocean-app-platform/)

## License

This GitHub Action and associated scripts and documentation in this project are released under the [MIT License](LICENSE).