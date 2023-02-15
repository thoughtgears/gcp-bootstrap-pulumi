package data

func RootUserGlobalRoles() []string {
	return []string{
		"roles/billing.admin",
		"roles/owner",
		"roles/billing.creator",
		"roles/resourcemanager.folderAdmin",
		"roles/resourcemanager.organizationAdmin",
		"roles/resourcemanager.projectCreator",
		"roles/resourcemanager.projectDeleter",
		"roles/billing.projectManager",
	}
}
