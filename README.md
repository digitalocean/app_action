# Deploy a [DigitalOcean App Platform](https://www.digitalocean.com/products/app-platform/) app using GitHub Actions.

 - Auto-deploy your app from source on commit, while allowing you to run tests or perform other operations before.
 - Auto-deploy your app from source and also update DigitalOcean Container Registry (DOCR) configuration in DigitalOcean [App Spec](https://docs.digitalocean.com/products/app-platform/references/app-specification-reference/) and deploy application with updated DOCR image.

**Note: This action only supports DOCR configuration changes for Auto-deploy**
# Usage

### DigitalOcean App Platform Auto-deploy with same app spec.
- Get DigitalOcean Personal Access token by following this [instructions](https://docs.digitalocean.com/reference/api/create-personal-access-token/).**(skip this step if you already have DigitalOcean Personal Access Token)**
- Declare DigitalOcean Personal Access Token as DIGITALOCEAN_ACCESS_TOKEN variable in the [secrets](https://docs.github.com/en/actions/reference/encrypted-secrets#creating-encrypted-secrets-for-a-repository) of github repository. 
- Add this step to deploy your application on DigitalOcean App Platform without changing any app spec configuration or making any other changes.
  ```yaml
  - name: DigitalOcean App Platform deployment
    uses: ParamPatel207/app_action@main
    with:
      app_name: my_DO_app
      token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
  ```
- DigitalOcean App Platform will now deploy your application.

### Update DigitalOcean Container Registry(DOCR) configuration of multiple component in App Spec

- Get DigitalOcean Personal Access token by following this [instructions](https://docs.digitalocean.com/reference/api/create-personal-access-token/)**(skip this step if you already have DigitalOcean Personal Access Token)**
- Declare DigitalOcean Personal Access Token as DIGITALOCEAN_ACCESS_TOKEN variable in the [secrets](https://docs.github.com/en/actions/reference/encrypted-secrets#creating-encrypted-secrets-for-a-repository) of github repository. 
- Add this step to update DigitalOcean Container Registry configuration of single or multiple [component]((https://www.digitalocean.com/blog/build-component-based-apps-with-digitalocean-app-platform/)) in app_spec
  ```yaml
  - name: DigitalOcean App Platform deployment
    uses: ParamPatel207/app_action@main
    with:
      app_name: my_DO_app
      token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
      list_of_image: '[
                        {
                          "name": " ",
                          "repository": " ",
                          "tag": ""
                        },
                        {
                          "name": " ",
                          "repository": " ",
                          "tag": " "
                        },
                      ]'
  ```
- DigitalOcean App Platform will now update your DOCR information in App Spec and then deploy your application.

**Note: Always use unique tag names to push image to the DigitalOcean Container Registry. This will allow you to deploy your application without delay. [ref](https://docs.digitalocean.com/products/container-registry/quickstart/)**

# Inputs
- `app_name` - Name of the app on App Platform.
- `list_of_image` - (optional)List of json object for providing information about name, repository and tag of the image in docr.(by default latest tag is used)
    ```json
    {
      "name": " ",
      "repository": " ",
      "tag": ""
    }
    ```
    - `name` - name of the component in [App Spec](https://docs.digitalocean.com/products/app-platform/references/app-specification-reference/)
    - `repository` - name of the DOCR repository with the following format- registry.digitalocean.com/<my-registry>/<my-image>
    - `tag` - tag of the image provided while pushing to DOCR (by default latest tag is used). 
    **We suggest always use unique tag value)**
- `token` - doctl authentication token (generate token by following this [instructions](https://docs.digitalocean.com/reference/api/create-personal-access-token/)

## Example:
Update DigitalOcean Container Registry(DOCR) configuration of single component in App Spec [example](https://github.com/ParamPatel207/docr_sample)

DigitalOcean App Platform Auto-deploy with same app spec. [example](https://github.com/ParamPatel207/sample_golang_github_action)

## Resources to know more about DigitalOcean App Platform App Spec
- [App Platform Guided App Spec Declaration](https://www.digitalocean.com/community/tech_talks/defining-your-app-specification-on-digitalocean-app-platform)
- [App Platform App Spec Blog](https://docs.digitalocean.com/products/app-platform/references/app-specification-reference/)
- [App Platform App Spec Components](https://www.digitalocean.com/blog/build-component-based-apps-with-digitalocean-app-platform/)

## Note for handling DigitalOcean Container Registry images: 
Because image manifests are cached in different regions, there may be a maximum delay of one hour between pushing to a tag that already exists in your registry and being able to pull the new image by tag. This may happen, for example, when using the :latest tag. To avoid the delay, use:
- Unique tags (other than :latest)
- SHA hash of Github commit
- SHA hash of the new manifest

## License
This GitHub Action and associated scripts and documentation in this project are released under the [MIT License](LICENSE).
