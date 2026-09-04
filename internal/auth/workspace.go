package auth

// Which workspace a repository's pipeline commands act in.
//
// This is CLI state, not platform state and not repository state, and it is
// held here for a reason that matters to the customer's repository: the
// declaration contract is "a chart and one `.asgard-pipeline.yaml`", so a
// platform identifier written into the checkout would be a third file the
// contract does not have. The platform does not hold it either - a workspace
// does not know which of your directories you keep its repository in.
//
// So it is recorded per machine, keyed by profile and by the repository's
// "owner/name". Keyed by profile because the same repository is legitimately
// bound in dev and in prod at once, and a single mapping would send a command
// meant for one to the other.
//
// An FDE-scaffolded repository that already records `workspace.id` in its
// `.asgard-config.json` is read from there instead: that file is committed, so
// a teammate cloning it does not have to be told again.

// RepoWorkspaces maps a repository's "owner/name" to a workspace id.
type RepoWorkspaces map[string]string

// WorkspaceFor returns the workspace bound to a repository under a profile.
func (s *Settings) WorkspaceFor(profile, repoFullName string) (string, bool) {
	byRepo, ok := s.Workspaces[profile]
	if !ok {
		return "", false
	}
	id, ok := byRepo[repoFullName]
	return id, ok && id != ""
}

// BindWorkspace records the workspace a repository's commands act in.
func (s *Settings) BindWorkspace(profile, repoFullName, workspaceID string) {
	if s.Workspaces == nil {
		s.Workspaces = map[string]RepoWorkspaces{}
	}
	if s.Workspaces[profile] == nil {
		s.Workspaces[profile] = RepoWorkspaces{}
	}
	s.Workspaces[profile][repoFullName] = workspaceID
}

// UnbindWorkspace forgets a repository's binding and reports whether there was
// one.
func (s *Settings) UnbindWorkspace(profile, repoFullName string) bool {
	byRepo, ok := s.Workspaces[profile]
	if !ok {
		return false
	}
	if _, ok := byRepo[repoFullName]; !ok {
		return false
	}
	delete(byRepo, repoFullName)
	if len(byRepo) == 0 {
		delete(s.Workspaces, profile)
	}
	return true
}

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
