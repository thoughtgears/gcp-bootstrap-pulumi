package data

type Data struct {
	Name                       string                      `yaml:"bootstrapProjectName"`
	OrganizationId             string                      `yaml:"organisationId"`
	OrganizationName           string                      `yaml:"organizationName"`
	RootUser                   string                      `yaml:"rootUser"`
	BillingAccount             string                      `yaml:"billingAccount"`
	OrgPolicies                []OrgPolicy                 `yaml:"organisationPolicies"`
	Folders                    []string                    `yaml:"folders,omitempty"`
	WorkloadIdentityFederation *WorkloadIdentityFederation `yaml:"workloadIdentityFederation,omitempty"`
}

type OrgPolicy struct {
	Policy string `yaml:"policy"`
	Value  bool   `yaml:"value"`
}

type WorkloadIdentityFederation struct {
	Providers []WorkloadIdentityFederationProvider `yaml:"providers"`
}

type WorkloadIdentityFederationProvider struct {
	Name                string `yaml:"name"`
	IssuerUrl           string `yaml:"issuerUrl"`
	AttributeIamSetting struct {
		Mapping string `yaml:"mapping"`
		Value   string `yaml:"value"`
	} `yaml:"attributeIamSetting"`
	AttributeMapping []struct {
		Key   string `yaml:"key"`
		Value string `yaml:"value"`
	} `yaml:"attributeMapping"`
}
