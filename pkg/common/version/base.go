package version

// Base version information.
//
// This is the fallback data used when version information from git is not
// provided via go ldflags. It provides an approximation of the Kubernetes
// version for ad-hoc builds (e.g. `go build`) that cannot get the version
// information from git.
//
// If you are looking at these fields in the git tree, they look
// strange. They are modified on the fly by the build process. The
// in-tree values are dummy values used for "git archive", which also
// works for GitHub tar downloads.
//
// When releasing a new Kubernetes version, this file is updated by
// build/mark_new_version.sh to reflect the new version, and then a
// git annotated tag (using format vX.Y where X == Major version and Y
// == Minor version) is created to point to the commit that updates
var (
	// TODO: Deprecate gitMajor and gitMinor, use only gitVersion
	// instead. First step in deprecation, keep the fields but make
	// them irrelevant. (Next we'll take it out, which may muck with
	// scripts consuming the kubectl version output - but most of
	// these should be looking at gitVersion already anyways.)
	gitMajor string = "" // major version, always numeric
	gitMinor string = "" // minor version, numeric possibly followed by "+"

	// semantic version, derived by build scripts (see
	// https://github.com/kubernetes/sig-release/blob/master/release-engineering/versioning.md#kubernetes-release-versioning
	// https://kubernetes.io/releases/version-skew-policy/
	// for a detailed discussion of this field)
	//
	// TODO: This field is still called "gitVersion" for legacy
	// reasons. For prerelease versions, the build metadata on the
	// semantic version is a git hash, but the version itself is no
	// longer the direct output of "git describe", but a slight
	// translation to be semver compliant.

	// NOTE: The $Format strings are replaced during 'git archive' thanks to the
	// companion .gitattributes file containing 'export-subst' in this same
	// directory.  See also https://git-scm.com/docs/gitattributes
	gitVersion   string = "latest"
	gitCommit    string = "" // sha1 from git, output of $(git rev-parse HEAD)
	gitTreeState string = ""            // state of git tree, either "clean" or "dirty"

	buildDate string = "1970-01-01T00:00:00Z" // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)