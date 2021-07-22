# Getting Started

These steps will get this sample Dockerfile application running for you using DigitalOcean App Platform.

**Note: Following these steps may result in charges for the use of DigitalOcean services**

## Requirements

* You need a DigitalOcean account. If you don't already have one, you can sign up at https://cloud.digitalocean.com/registrations/new
    
## Deploying the App

Click this button to deploy the app to the DigitalOcean App Platform. If you are not logged in, you will be prompted to log in with your DigitalOcean account.

[![Deploy to DigitalOcean](https://www.deploytodo.com/do-btn-blue.svg)](https://cloud.digitalocean.com/apps/new?repo=https://github.com/digitalocean/sample-dockerfile/tree/main)

Using this button disable the "Auto deploy changes on push" feature as you are using this repo directly. If you wish to try that feature, you will need to make your own copy of this repository.

To make a copy, click the Fork button above and follow the on-screen instructions. In this case, you'll be forking this repo as a starting point for your own app (see [GitHub documentation](https://docs.github.com/en/github/getting-started-with-github/fork-a-repo) to learn more about forking repos.

After forking the repo, you should now be viewing this README in your own github org (e.g. `https://github.com/<your-org>/sample-dockerfile`). To deploy the new repo, visit https://cloud.digitalocean.com/apps and click "Create App" or "Launch Your App". Then, select the repository you created and be sure to select the `main` branch.

After clicking the "Deploy to DigitalOcean" button or completing the instructions above to fork the repo, follow these steps:

1. Select which region you wish to deploy your app to and click Next. The closest region to you should be selected by default. All App Platform apps are routed through a global CDN so this will not affect your app performance, unless it needs to talk to external services.
1. On the following screen, leave all the fields as they are and click Next.
1. Confirm your Plan settings and how many containers you want to launch and click **Launch Basic/Pro App**.
1. You should see a "Building..." progress indicator. And you can click "Deployments"→"Details" to see more details of the build.
1. It can take a few minutes for the build to finish, but you can follow the progress by clicking the "Details" link in the top banner.
1. Once the build completes successfully, click the "Live App" link in the header and you should see your running application in a new tab, displaying the home page.

## Making Changes to Your App

If you followed the steps above to fork the repo and used your own copy when deploying the app, you can enjoy automatic deploys whenever changes are made to the repo. During these automatic deployments, your application will never pause or stop serving request because App Platform offers zero-downtime deployments.

Here's an example code change you can make for this app:

1. Edit `main.go` and replace the "Hello!" greeting on line 31 with a different greeting.
1. Commit the change to the `main` branch. Normally it's a better practice to create a new branch for your change and then merge that branch to `main` after review, but for this demo you can commit to the `main` branch directly.
1. Visit https://cloud.digitalocean.com/apps and navigate to your sample app.
1. You should see a "Building..." progress indicator, just like when you first created the app.
1. Once the build completes successfully, click the "Live App" link in the header and you should see your updated application running. You may need to force refresh the page in your browser (e.g. using Shift+Reload).

## Learn More

You can learn more about the App Platform and how to manage and update your application at https://www.digitalocean.com/docs/app-platform/.

## Deleting the App

When you no longer need this sample application running live, you can delete it by following these steps:
1. Visit the Apps control panel at https://cloud.digitalocean.com/apps
1. Navigate to the sample app
1. Choose "Settings"->"Destroy"

**Note: If you don't delete your app, charges for the use of DigitalOcean services will continue to accrue.**