package etc_meisai

// Version is the current version of the etc_meisai module
const Version = "v0.0.3"

// GetEtcMeisaiVersion returns the version of the etc_meisai module
func GetEtcMeisaiVersion() (string, error) {
	return Version, nil
}