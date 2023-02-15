package data

func ServiceAccountGlobalRoles() []string {
	return []string{
		"roles/resourcemanager.folderAdmin",
		"roles/logging.configWriter",
		"roles/resourcemanager.organizationViewer",
		"roles/resourcemanager.projectDeleter",
		"roles/resourcemanager.projectCreator",
		"roles/orgpolicy.policyAdmin",
		"roles/iam.serviceAccountAdmin",
		"roles/iam.securityAdmin",
		"roles/billing.projectManager",
	}
}
