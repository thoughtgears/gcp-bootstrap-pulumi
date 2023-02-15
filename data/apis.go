package data

func Apis() []string {
	// Apis that needs to be enabled initially to enable IAM, logging, resources and other.
	return []string{
		"cloudresourcemanager.googleapis.com",
		"serviceusage.googleapis.com",
		"iam.googleapis.com",
		"logging.googleapis.com",
		"securitycenter.googleapis.com",
		"iamcredentials.googleapis.com",
		"sts.googleapis.com",
	}
}
