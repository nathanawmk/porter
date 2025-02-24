package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/manifest"

	"get.porter.sh/porter/pkg/secrets"

	"github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_applyDefaultOptions(t *testing.T) {
	p := NewTestPorter(t)
	err := p.Create()
	require.NoError(t, err)

	opts := InstallOptions{
		&BundleActionOptions{
			sharedOptions: sharedOptions{
				bundleFileOptions: bundleFileOptions{
					File: "porter.yaml",
				},
			},
		},
	}
	err = opts.Validate([]string{}, p.Porter)
	require.NoError(t, err)

	p.Debug = true
	err = p.applyDefaultOptions(&opts.sharedOptions)
	require.NoError(t, err)

	assert.NotNil(t, p.Manifest, "Manifest should be loaded")
	assert.NotEqual(t, &manifest.Manifest{}, p.Manifest, "Manifest should not be empty")
	assert.Equal(t, p.Manifest.Name, opts.Name, "opts.Name should be set using the available manifest")
}

func TestInstallOptions_validateInstallationName(t *testing.T) {
	testcases := []struct {
		name      string
		args      []string
		wantClaim string
		wantError string
	}{
		{"none", nil, "", ""},
		{"name set", []string{"wordpress"}, "wordpress", ""},
		{"too many args", []string{"wordpress", "extra"}, "", "only one positional argument may be specified, the installation name, but multiple were received: [wordpress extra]"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewInstallOptions()
			err := opts.validateInstallationName(tc.args)

			if tc.wantError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.wantClaim, opts.Name)
			} else {
				require.EqualError(t, err, tc.wantError)
			}
		})
	}
}

func TestInstallOptions_validateDriver(t *testing.T) {
	testcases := []struct {
		name       string
		driver     string
		wantDriver string
		wantError  string
	}{
		{"debug", "debug", DebugDriver, ""},
		{"docker", "docker", DockerDriver, ""},
		{"invalid driver provided", "dbeug", "", "unsupported driver or driver not found in PATH: dbeug"},
	}

	cxt := context.NewTestContext(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := InstallOptions{
				&BundleActionOptions{
					sharedOptions: sharedOptions{
						Driver: tc.driver,
					},
				},
			}
			err := opts.validateDriver(cxt.Context)

			if tc.wantError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.wantDriver, opts.Driver)
			} else {
				require.EqualError(t, err, tc.wantError)
			}
		})
	}
}

func TestPorter_InstallBundle_WithDepsFromTag(t *testing.T) {
	p := NewTestPorter(t)

	cacheDir, _ := p.Cache.GetCacheDir()
	p.TestConfig.TestContext.AddTestDirectory("testdata/cache", cacheDir)

	// Make some fake credentials to give to the install operation, they won't be used because it's a dummy driver
	cs := credentials.NewCredentialSet("wordpress",
		valuesource.Strategy{
			Name: "kubeconfig",
			Source: valuesource.Source{
				Key:   secrets.SourceSecret,
				Value: "kubeconfig",
			},
		})
	p.TestCredentials.TestSecrets.AddSecret("kubeconfig", "abc123")
	err := p.Credentials.Save(cs)
	require.NoError(t, err, "Credentials.Save failed")

	opts := NewInstallOptions()
	opts.Driver = DebugDriver
	opts.Reference = "localhost:5000/wordpress:v0.1.3"
	opts.CredentialIdentifiers = []string{"wordpress"}
	opts.Params = []string{"wordpress-password=mypassword"}
	err = opts.Validate(nil, p.Porter)
	require.NoError(t, err, "Validate install options failed")

	err = p.InstallBundle(opts)
	require.NoError(t, err, "InstallBundle failed")
}
