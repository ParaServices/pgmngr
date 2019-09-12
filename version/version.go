package version

// AppRevision is the git SHA of current release. The value is injected by the
// build pipeline.
const AppRevision = ""

// AppRevisionTag is the git tag of the current release. The value is injected
// by the build pipeline, if known.
const AppRevisionTag = ""

// AppRevisionOrTag ...
func AppRevisionOrTag() string {
	if AppRevisionTag == "" {
		return AppRevision
	}
	return AppRevisionTag
}
