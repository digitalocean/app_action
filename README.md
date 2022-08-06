# Deploy a [DigitalOcean App Platform](https://www.digitalocean.com/products/app-platform/) app using GitHub Actions.

 - Auto-deploy your app from source on commit, while allowing you to run tests or perform other operations before.
 - Auto-deploy your app from source and also update DockerHub / DigitalOcean Container Registry (DOCR) configuration in DigitalOcean [App Spec](https://docs.digitalocean.com/products/app-platform/reference/app-spec/) and deploy application with updated container image.


# Usage
### Deploy via GH Action and let DigitalOcean App Platform build and deploy your app.
- Get DigitalOcean Personal Access token by following this [instructions](https://docs.digitalocean.com/reference/api/create-personal-access-token/).**(skip this step if you already have DigitalOcean Personal Access Token)**
- Declare DigitalOcean Personal Access Token as DIGITALOCEAN_ACCESS_TOKEN variable in the [secrets](https://docs.github.com/en/actions/reference/encrypted-secrets#creating-encrypted-secrets-for-a-repository) of github repository. 
- [Create a GitHub Action workflow file](https://docs.github.com/en/actions/learn-github-actions/introduction-to-github-actions#create-an-example-workflow) and add this step below to it or add this to your existing action.
  ```yaml
  - name: DigitalOcean App Platform deployment
    uses: digitalocean/app_action@main
    with:
      app_name: my_DO_app
      token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
  ```
- This step will trigger a deploy to your App on DigitalOcean App Platform

### Deploy an one or more app components from a DigitalOcean Container Registry (DOCR) or DockerHub

- Get DigitalOcean Personal Access token by following this [instructions](https://docs.digitalocean.com/reference/api/create-personal-access-token/)**(skip this step if you already have DigitalOcean Personal Access Token)**
- Declare DigitalOcean Personal Access Token as DIGITALOCEAN_ACCESS_TOKEN variable in the [secrets](https://docs.github.com/en/actions/reference/encrypted-secrets#creating-encrypted-secrets-for-a-repository) of github repository. 
- Add this step to update DigitalOcean Container Registry configuration of single or multiple [component]((https://www.digitalocean.com/blog/build-component-based-apps-with-digitalocean-app-platform/)) in app_spec
  ```yaml
  - name: DigitalOcean App Platform deployment
    uses: digitalocean/app_action@v1.0.0 # replace this with current version from https://github.com/digitalocean/app_action/releases
    with:
      app_name: my_DO_app
      token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
      images: '[
                {
                  "name": "sample-golang",
                   "image":{
                    "registry_type": "DOCR",
                    "repository": "add_sample",
                    "tag": "a5cae3e"
                  },
                },
                {
                  "name": "sample-add",
                  "image":{
                    "registry_type": "DOCKER_HUB",
                    "registry": "nginxdemos",
                    "repository": "hello",
                    "tag": "0.2"
                  },
              ]'
  ```
- DigitalOcean App Platform will now update your container image information in App Spec and then deploy your application.
- This step will trigger a DigitalOcean App Platform deployment of your app using the images specified.

**Note: Always use unique tag names (i.e. `1.2.3` instead of `latest`) to push image to the DigitalOcean Container Registry. This will allow you to deploy your application without delay. [ref](https://docs.digitalocean.com/products/container-registry/quickstart/)**

# Inputs
- `app_name` - Name of the app on App Platform.
- `images` - (optional)List of json objects (of type ImageSourceSpec, see [Reference for App Specification](https://docs.digitalocean.com/products/app-platform/references/app-specification-reference/)) for providing information about name, registry type, repository, and tag of the image in .
    ```json
    [{
      "name": " ", #component name
      "image":{ 
        "registry_type": "DOCKER_HUB", #Registry type, DOCR and DOCKER_HUB are supported
        "registry": "nginxdemos", # DockerHub only, the registry name
        "repository": "hello", # repository name
        "tag": "0.2" # tag name
      },
    }]
    ```
    - `name` - name of the component in [App Spec](https://docs.digitalocean.com/products/app-platform/references/app-specification-reference/)
    - `repository` - name of the DOCR repository with the following format- registry.digitalocean.com/<my-registry>/<my-image>
    - `tag` - tag of the image provided while pushing to DOCR (by default latest tag is used). 
    **We suggest always use unique tag value)**
- `token` - doctl authentication token (generate token by following this [instructions](https://docs.digitalocean.com/reference/api/create-personal-access-token/)

## Example:
Update DigitalOcean container image configuration of single component in App Spec [example](https://github.com/digitalocean/sample-golang-docr-github-action)

DigitalOcean App Platform Auto-deploy with same app spec. [example](https://github.com/digitalocean/sample-golang-github-action)

## Resources to know more about DigitalOcean App Platform App Spec
- [App Platform Guided App Spec Declaration](https://www.digitalocean.com/community/tech_talks/defining-your-app-specification-on-digitalocean-app-platform)
- [App Platform App Spec Blog](https://docs.digitalocean.com/products/app-platform/references/app-specification-reference/)
- [App Platform App Spec Components](https://www.digitalocean.com/blog/build-component-based-apps-with-digitalocean-app-platform/)

## Note for handling DigitalOcean Container Registry images: 
Because image manifests are cached in different regions, there may be a maximum delay of one hour between pushing to a tag that already exists in your registry and being able to pull the new image by tag. This may happen, for example, when using the :latest tag. To avoid the delay, use:
- Unique tags (other than :latest)
- SHA hash of Github commit
- SHA hash of the new manifest

## Development

- Install gomock with `go install github.com/golang/mock/mockgen@v1.6.0`
- `go generate ./...` to generate the mocks

## License
This GitHub Action and associated scripts and documentation in this project are released under the [MIT License](LICENSE).
