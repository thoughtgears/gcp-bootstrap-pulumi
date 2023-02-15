package main

import (
	"fmt"
	"gcp-bootstrap/data"
	"log"
	"os"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/iam"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/serviceaccount"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/projects"

	"github.com/pulumi/pulumi-random/sdk/v4/go/random"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gopkg.in/yaml.v3"
)

func main() {
	var d data.Data

	dataFile, err := os.ReadFile("data.yaml")
	if err != nil {
		log.Fatal(err)
	}

	if err := yaml.Unmarshal(dataFile, &d); err != nil {
		log.Fatal(err)
	}

	/*
		Create a new landing-zone project for GCP
		This will create the initial project, activate the required API:s and create a
		service account to use for further management of google cloud.
		If defined it will also set up a workload identity federation pool with the providers
		specified in the d file.
		It will also set root user permissions on the organization to manage projects and the
		organization. The root user should be a secure user and not a personal account.
	*/
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Ensure you have a proper suffix ID at the end of creating a project
		projectSuffix, err := random.NewRandomInteger(ctx, "project_suffix", &random.RandomIntegerArgs{
			Min: pulumi.Int(10000),
			Max: pulumi.Int(99999),
		})
		if err != nil {
			return err
		}

		// Create the new landing-zone project to have as a baseline when using pulumi to be able to create other
		// projects in your GCP organization.
		project, err := organizations.NewProject(ctx, d.Name, &organizations.ProjectArgs{
			Name:              pulumi.String(d.Name),
			BillingAccount:    pulumi.String(d.BillingAccount),
			AutoCreateNetwork: pulumi.Bool(false),
			OrgId:             pulumi.String(d.OrganizationId),
			ProjectId:         pulumi.Sprintf("%s-%v", d.Name, projectSuffix.Result),
		})
		if err != nil {
			return err
		}

		// Enable project API:s for landing-zone, all apis can be found in data.Apis() function.
		for _, api := range data.Apis() {
			_, err := projects.NewService(ctx, api, &projects.ServiceArgs{
				DisableDependentServices: pulumi.Bool(false),
				Project:                  project.ProjectId,
				Service:                  pulumi.String(api),
			})
			if err != nil {
				return err
			}
		}

		// Create the pulumi service account for interactions with all pulumi creation resources on
		// a global level, such as new projects and other global resources.
		account, err := serviceaccount.NewAccount(ctx, "pulumi", &serviceaccount.AccountArgs{
			AccountId:   pulumi.String("pulumi"),
			DisplayName: pulumi.String("Pulumi Service Account"),
			Description: pulumi.String("Service account to use with pulumi deployments"),
			Project:     project.ProjectId,
		})
		if err != nil {
			return err
		}

		// Ensure the new SA is owner of the landing-zone account
		if _, err := projects.NewIAMBinding(ctx, "service-account-owner", &projects.IAMBindingArgs{
			Members: pulumi.StringArray{pulumi.Sprintf("serviceAccount:%s", account.Email)},
			Project: project.ProjectId,
			Role:    pulumi.String("roles/owner"),
		}); err != nil {
			return err
		}

		// Set service account global permissions on organization
		for _, role := range data.ServiceAccountGlobalRoles() {
			if _, err := organizations.NewIAMMember(ctx, fmt.Sprintf("service-account-%s", role), &organizations.IAMMemberArgs{
				Member: pulumi.Sprintf("serviceAccount:%s", account.Email),
				OrgId:  pulumi.String(d.OrganizationId),
				Role:   pulumi.String(role),
			}); err != nil {
				return err
			}
		}

		// Set root user permissions both on the service account to assume it but also on the organization level
		if _, err := serviceaccount.NewIAMMember(ctx, "pulumi-service-account-token-creator", &serviceaccount.IAMMemberArgs{
			Member:           pulumi.String(fmt.Sprintf("user:%s", d.RootUser)),
			Role:             pulumi.String("roles/iam.serviceAccountTokenCreator"),
			ServiceAccountId: account.ID(),
		}); err != nil {
			return err
		}

		if _, err := serviceaccount.NewIAMMember(ctx, "pulumi-service-account-user", &serviceaccount.IAMMemberArgs{
			Member:           pulumi.String(fmt.Sprintf("user:%s", d.RootUser)),
			Role:             pulumi.String("roles/iam.serviceAccountUser"),
			ServiceAccountId: account.ID(),
		}); err != nil {
			return err
		}

		for _, role := range data.RootUserGlobalRoles() {
			if _, err := organizations.NewIAMMember(ctx, fmt.Sprintf("root-user-%s", role), &organizations.IAMMemberArgs{
				Member: pulumi.String(fmt.Sprintf("user:%s", d.RootUser)),
				OrgId:  pulumi.String(d.OrganizationId),
				Role:   pulumi.String(role),
			}); err != nil {
				return err
			}
		}

		// Impersonates service account after landing zone creation, will use the pulumi service account
		// to execute the rest of the infrastructure to become owner of it. It also needed to be a service
		// account to update certain resources in the org.
		if _, err := local.NewCommand(ctx, "impersonate-service-account", &local.CommandArgs{
			Update: pulumi.Sprintf(`gcloud config set auth/impersonate_service_account %s`, account.Email),
		}); err != nil {
			return err
		}

		// Set the pulumi config with the newly created project
		if _, err := local.NewCommand(ctx, "set-project-config", &local.CommandArgs{
			Update: pulumi.Sprintf(`pulumi config set gcp:project %s`, project.ID()),
		}); err != nil {
			return err
		}

		// If any org policies are defined, set all the org polices defined in the d file.
		// Needs to be in a list of maps with policy name and value for each of the policies.
		if d.OrgPolicies != nil {
			for _, policy := range d.OrgPolicies {
				_, err := organizations.NewPolicy(ctx, policy.Policy, &organizations.PolicyArgs{
					BooleanPolicy: &organizations.PolicyBooleanPolicyArgs{
						Enforced: pulumi.Bool(policy.Value),
					},
					Constraint: pulumi.String(policy.Policy),
					OrgId:      pulumi.String(d.OrganizationId),
				})
				if err != nil {
					return err
				}
			}
		}

		// If any folders are defined, set the folders defined in the d file.
		// TODO: enable nested folders from d file
		if d.Folders != nil {
			for _, folder := range d.Folders {
				if _, err := organizations.NewFolder(ctx, folder, &organizations.FolderArgs{
					DisplayName: pulumi.String(folder),
					Parent:      pulumi.String(fmt.Sprintf("organizations/%s", d.OrganizationId)),
				}); err != nil {
					return nil
				}
			}
		}

		// If workload identity federation is defined or enabled, set up a pool and all
		// the providers defined in the d file.

		// Create the new pulumi-landing-zone pool to enable providers to use the service account.
		pool, err := iam.NewWorkloadIdentityPool(ctx, "pulumi-landinzone", &iam.WorkloadIdentityPoolArgs{
			WorkloadIdentityPoolId: pulumi.String("pulimi-landingzone"),
			Project:                project.ProjectId,
		})
		if err != nil {
			return err
		}

		var am pulumi.StringMap
		am = make(map[string]pulumi.StringInput)

		// Ensure we loop over all our providers
		for _, provider := range d.WorkloadIdentityFederation.Providers {
			// Create a attribute map to use with the provider based on the data in the yaml file.
			for _, attributeMap := range provider.AttributeMapping {
				am[attributeMap.Key] = pulumi.String(attributeMap.Value)
			}

			p, err := iam.NewWorkloadIdentityPoolProvider(ctx, provider.Name, &iam.WorkloadIdentityPoolProviderArgs{
				WorkloadIdentityPoolId:         pool.WorkloadIdentityPoolId,
				WorkloadIdentityPoolProviderId: pulumi.String(provider.Name),
				AttributeMapping:               am,
				Project:                        project.ProjectId,
				Oidc: &iam.WorkloadIdentityPoolProviderOidcArgs{
					IssuerUri: pulumi.String(provider.IssuerUrl),
				},
			})
			if err != nil {
				return err
			}

			ctx.Export(fmt.Sprintf("workload-identity provider: %s", provider.Name), p.ID())

			// Set the Workload provider to interact with the provider through our pulumi SA
			if _, err := serviceaccount.NewIAMMember(ctx, provider.AttributeIamSetting.Value, &serviceaccount.IAMMemberArgs{
				Member:           pulumi.Sprintf("principalSet://iam.googleapis.com/%s/%s/%s}", pool.Name, provider.AttributeIamSetting.Mapping, provider.AttributeIamSetting.Value),
				Role:             pulumi.String("roles/iam.workloadIdentityUser"),
				ServiceAccountId: account.ID(),
			}, pulumi.DependsOn([]pulumi.Resource{p})); err != nil {
				return err
			}
		}

		// Generate Pulumi readme file
		readme, err := os.ReadFile("./README.md")
		if err != nil {
			return fmt.Errorf("could not read README.md: %w", err)
		}

		ctx.Export("pulumi service account", account.Email)
		ctx.Export("workload identity pool", pool.ID())
		ctx.Export("landingzone", project.ProjectId)
		ctx.Export("readme", pulumi.String(readme))

		return nil
	})
}
