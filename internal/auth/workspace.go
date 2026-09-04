package auth

// The workspace to use where no checkout answers the question.
//
// A checkout's own workspace is recorded in its `.asgard-cli.yaml`, beside the
// declaration it belongs to, and committed - see internal/binding for why it
// lives there rather than here. What is left for this file is the case that has
// no checkout at all: `workspace list` run from anywhere, a first look before
// cloning anything.
//
// There is deliberately no per-repository override here. One existed, and it
// was a per-machine setting silently outranking a committed one, which is
// exactly how "why did it deploy there" happens three weeks later. Overriding
// is --workspace and ASGARD_WORKSPACE, both of which are visible at the moment
// they apply and gone afterwards.

// FallbackWorkspace returns the workspace to use when no repository is in play
// - listing pipelines from anywhere, for instance.
func (s *Settings) FallbackWorkspace(profile string) (string, bool) {
	id, ok := s.DefaultWorkspaces[profile]
	return id, ok && id != ""
}

// SetFallbackWorkspace records the workspace for commands run outside a
// repository.
func (s *Settings) SetFallbackWorkspace(profile, workspaceID string) {
	if s.DefaultWorkspaces == nil {
		s.DefaultWorkspaces = map[string]string{}
	}
	s.DefaultWorkspaces[profile] = workspaceID
}
