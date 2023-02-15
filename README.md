# Bootstrap for GCP landing-zone

This will boostrap a landing-zone for your GCP environment. You will get the initial account setup with a pulumi SA  
to be able to use with creating more projects and managing them with pulumi.  

To start you should have created an organization in Google Cloud and gone through the initial steps such as  
setting up a billing account and added your first user.  Your user will need to have Billing Account Administrator,  
Project Creator and Project Billing Manager. After this initial bootstrap it's recommended to use the pulumi  
service account as an actor when modifying your infrastructure.  

Initial run of the bootstrap has to be done locally with your own user, after the whole stack is created you can  
use GHA to update your landing-zone. You will need to set a few variables in your repo.  

* GOOGLE_PROVIDER - projects/1234/locations/global/workloadIdentityPools/pool-name/providers/provider-name
* GOOGLE_SERVICE_ACCOUNT - service account email
* PULUMI_WORKING_DIRECTORY - ./ (or any other directory you run pulumi from)

You will also have to add a secret in your repo: 

* PULUMI_ACCESS_TOKEN - your token generated from pulumi site

## Data file

To work properly you need a data file for your initial bootstrap.  
You also need to populate it with the following

```yaml
organisationId: "orgId"
organisationName: "orgName"
bootstrapProjectName: "landingzone"
billingAccount: "billingID"
rootUser: "your initial root user"
organisationPolicies:
  - policy: "constraints/compute.skipDefaultNetworkCreation"
    value: true
  - policy: "constraints/iam.disableServiceAccountKeyCreation"
    value: true
folders:
  - sandbox
  - production
workloadIdentityFederation:
  providers:
    - name: "github-actions-gcp-bootstrap"
      issuerUrl: "https://token.actions.githubusercontent.com"
      attributeIamSetting:
        mapping: "attribute.repository"
        value: "your bootstrap repo"
      attributeMapping:
        - key: "google.subject"
          value: "assertion.sub"
        - key: "attribute.actor"
          value: "assertion.actor"
        - key: "attribute.repository"
          value: "assertion.repository"
````

#### Useful commands to know before bootstrapping.

`gcloud auth login` **logs your gcloud cli into google cloud and you are ready to use the SDK**  
`gcloud auth application-default login` **logs your gcloud in and gets you credentials for the ADC that google sdks often need**  
`gcloud --impersonate-service-account=<name>@<project_id>.iam.gserviceaccount.com <gcloud command>` **impersonate the service account while running commands**  
`gcloud config set auth/impersonate_service_account <name>@<project_id>.iam.gserviceaccount.com` **set impersonation in gcloud config**  
`gcloud config unset auth/impersonate_service_account` **unset impersonation of service account in gcloud config**  