# DigitalOcean App Platform Image and DigitalOcean Registry publish
This action can be either used to deploy to digitalocean app platform using github action or can be used to update docr images in digitalocean app platform App Spec.(https://docs.digitalocean.com/products/app-platform/references/app-specification-reference/)
# Usage
Add this step to deploy your application on DigitalOcean App Platform using DigitalOcean Container Registry.

### Example:

Below example shows deployment to App Platform while updating Digital Ocean App Spec with update digitalocean container registry.
```yaml
    - name: DigitalOcean App Platform deployment
      uses: ParamPatel207/app_action@go_attempt
      with:
        app_name: App Platform Demo
        token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
        list_of_image: '[
                          {
                            "name": "frontend",
                            "repository": "registry.digitalocean.com/<my-registry>/<my-image>",
                            "tag": "latest"
                          },
                          {
                            "name": "landing",
                            "repository": "registry.digitalocean.com/<my-registry>/<my-image>",
                            "tag": "test1"
                          },
                          {
                            "name": "api",
                            "repository": "registry.digitalocean.com/<my-registry>/<my-image>",
                            "tag": "test2"
                          }
                        ]'
```
Sample golang application for deployment with docr update is as follows: https://github.com/ParamPatel207/docr_sample



Below example shows deployment to App Platform without updating the app spec.
```yaml
on:
  push:
    branches:
      - master
jobs:
  deploy:
    runs-on: ubuntu-latest
    name: Deploy App
    steps:
    - name: Checkout code
      uses: actions/checkout@v2
    - name: DigitalOcean App Platform deployment
      uses: ParamPatel207/app_action@go_attempt
      with:
        app_name: sample-golang
        token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
```
Sample golang application for no docr update deployment is as follows: https://github.com/ParamPatel207/sample_golang_github_action
# Inputs
- `app_name` - Name of the app on App Platform.
- `list_of_image` - (optional)List of json object for providing information about name,repository and tag of the image in docr.(By default tag of the image is latest)
- `token` - doctl authentication token(generate token by following https://docs.digitalocean.com/reference/api/create-personal-access-token/)

## License

This GitHub Action and associated scripts and documentation in this project are released under the [MIT License](LICENSE).
