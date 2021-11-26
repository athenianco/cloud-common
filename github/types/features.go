package types

const (
	// FeatureGHE is set for accounts based on Github Enterprise instance, instead of github.com.
	FeatureGHE = Feature("athenian.github.ghe")
	// FeatureNoConsistency disables consistency checks for an account.
	FeatureNoConsistency = Feature("athenian.github.no_consistency")
)
