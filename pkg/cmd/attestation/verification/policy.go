package verification

import (
	"encoding/hex"
	"fmt"

	"github.com/cli/cli/v2/pkg/cmd/attestation/artifact"

	"github.com/sigstore/sigstore-go/pkg/fulcio/certificate"
	"github.com/sigstore/sigstore-go/pkg/verify"
)

// represents the GitHub hosted runner in the certificate RunnerEnvironment extension
const GitHubRunner = "github-hosted"

// BuildDigestPolicyOption builds a verify.ArtifactPolicyOption
// from the given artifact digest and digest algorithm
func BuildDigestPolicyOption(a artifact.DigestedArtifact) (verify.ArtifactPolicyOption, error) {
	// sigstore-go expects the artifact digest to be decoded from hex
	decoded, err := hex.DecodeString(a.Digest())
	if err != nil {
		return nil, err
	}
	return verify.WithArtifactDigest(a.Algorithm(), decoded), nil
}

type EnforcementCriteria struct {
	Certificate   certificate.Summary
	PredicateType string
	SANRegex      string
	SAN           string
}

func (c EnforcementCriteria) Valid() error {
	if c.Certificate.Issuer == "" {
		return fmt.Errorf("Issuer must be set")
	}
	if c.Certificate.RunnerEnvironment != "" && c.Certificate.RunnerEnvironment != GitHubRunner {
		return fmt.Errorf("RunnerEnvironment must be set to either \"\" or %s", GitHubRunner)
	}
	if c.Certificate.SourceRepositoryOwnerURI == "" {
		return fmt.Errorf("SourceRepositoryOwnerURI must be set")
	}
	if c.PredicateType == "" {
		return fmt.Errorf("PredicateType must be set")
	}
	if c.SANRegex == "" && c.SAN == "" {
		return fmt.Errorf("SANRegex or SAN must be set")
	}
	return nil
}

func (c EnforcementCriteria) BuildPolicyInformation() string {
	template :=
		`
The following policy criteria will be enforced against all attestations:
- Attestation predicate type must match %s
- Attestation must be signed by a certificate whose OIDC issuer matches %s
- Attestation must be associated with an artifact built in an organization whose URI is %s`

	info := fmt.Sprintf(template, c.PredicateType, c.Certificate.Issuer, c.Certificate.SourceRepositoryOwnerURI)

	if c.Certificate.SourceRepositoryURI != "" {
		info += fmt.Sprintf("\n- Attestation must be associated with an artifact built in a repository whose URI is %s", c.Certificate.SourceRepositoryURI)
	}

	if c.Certificate.RunnerEnvironment == GitHubRunner {
		info += "\n- Attestation's signing certificate must be generated by an Action workflow executed in a GitHub hosted runner"
	}

	if c.SAN != "" {
		info += fmt.Sprintf("\n- Attestation's signing certificate must have a Subject Alternative Name matching the exact value %s", c.SAN)
	} else if c.SANRegex != "" {
		info += fmt.Sprintf("\n- Attestation's signing certificate must have a Subject Alternative Name matching the regex %s", c.SANRegex)
	}

	return info
}
