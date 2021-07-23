# DigitalOcean App Platform Image and DigitalOcean Container Registry publish
This action can be used to redeploy application on the DigitalOcean's [App Platform](https://www.digitalocean.com/products/app-platform/) using github action. This Action has two use cases one is to redeploy your application on App Platform with same configuration. The other use case is to update the DigitalOcean Container Registry configuration and deploy to App Platform. This github action uses DigitalOcean AppSpec [App Spec](https://docs.digitalocean.com/products/app-platform/references/app-specification-reference/).
# Usage
### DigitalOcean App Platform redeploy with same app spec.

Add this step to deploy your application on DigitalOcean App Platform without changing any app spec configuration or making any other changes.
```yaml
- name: DigitalOcean App Platform deployment
  uses: ParamPatel207/app_action@main
  with:
    app_name: 
    token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
```
DigitalOcean App Platform will now deploy your application.

### Update DigitalOcean Container Registry of multiple component in App Spec

Add this step to update single or multiple DigitalOcean Container Registry of each component in app_spec
```yaml
- name: DigitalOcean App Platform deployment
  uses: ParamPatel207/app_action@main
  with:
    app_name: 
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
DigitalOcean App Platform will now update your DOCR information in App Spec and then deploy your application.
(Please use unique tag value for DigitalOcean Container Registry Push instead of latest)

# Inputs
- `app_name` - Name of the app on App Platform.
- `list_of_image` - (optional)List of json object for providing information about name,repository and tag of the image in docr.(By default tag of the image is latest)
    ```json
    {
      "name": " ",
      "repository": " ",
      "tag": ""
    }
    ```
    - `name` - name of the component in [App Spec](https://docs.digitalocean.com/products/app-platform/references/app-specification-reference/)
    - `repository` - name of the DOCR repository with the following format registry.digitalocean.com/<my-registry>/<my-image>
    - `tag` - tag of the image provided while pushing to docr(by default its latest tag. We suggest always use unique tag value for any deployment)
- `token` - doctl authentication token (generate token by following https://docs.digitalocean.com/reference/api/create-personal-access-token/)

## Example:

Sample golang application for deployment with docr update. [example](https://github.com/ParamPatel207/docr_sample)

Sample golang application for redeployment. [example](https://github.com/ParamPatel207/sample_golang_github_action)

## Resources to know more about DigitalOcean App Platform App Spec
- [App Platform Guided App Spec Declaration](https://www.digitalocean.com/community/tech_talks/defining-your-app-specification-on-digitalocean-app-platform)
- [App Platform App Spec Blog](https://docs.digitalocean.com/products/app-platform/references/app-specification-reference/)
- [App Platform App Spec Components](https://www.digitalocean.com/blog/build-component-based-apps-with-digitalocean-app-platform/)
## Contributing



## License

This GitHub Action and associated scripts and documentation in this project are released under the [MIT License](LICENSE).
