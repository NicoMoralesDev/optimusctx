package release

type DistributionPolicy struct {
	ProductShape      ProductShape
	IntendedUsers     []string
	SupportedChannels []DistributionChannel
	Upgrade           UpgradePolicy
	Support           SupportPolicy
	DeferredScope     []DeferredScopeItem
}

type ProductShape struct {
	Summary                       string
	LocalFirst                    bool
	SingleBinary                  bool
	ManagedService                bool
	AutomaticConfigEditsByDefault bool
}

type DistributionChannel struct {
	ID                string
	Name              string
	UserFacingInstall string
	PublicationTarget string
	Audience          string
	UpgradePath       string
	RollbackPath      string
	SupportBoundary   string
}

type UpgradePolicy struct {
	VerificationCommands []string
	ArchiveExpectation   string
	PackageManagerRule   string
}

type SupportPolicy struct {
	SupportedCommands  []string
	IssueTracking      string
	OperatorModel      string
	ConfigMutationRule string
}

type DeferredScopeItem struct {
	Name   string
	Reason string
}

func CurrentDistributionPolicy() DistributionPolicy {
	return DistributionPolicy{
		ProductShape: ProductShape{
			Summary:                       "OptimusCtx ships as a local-first single binary with explicit, operator-controlled configuration.",
			LocalFirst:                    true,
			SingleBinary:                  true,
			ManagedService:                false,
			AutomaticConfigEditsByDefault: false,
		},
		IntendedUsers: []string{
			"Developers and coding-agent operators who are comfortable with CLI tooling, local PATH management, and explicit MCP setup.",
			"Teams evaluating the existing local-first runtime on their own machines instead of expecting a hosted onboarding path.",
		},
		SupportedChannels: []DistributionChannel{
			{
				ID:                "github-release-archive",
				Name:              "GitHub Release archives",
				UserFacingInstall: "Download the tagged archive from GitHub Releases, unpack it, place `optimusctx` on your PATH, then verify with `optimusctx version`, `optimusctx status`, and `optimusctx doctor`.",
				PublicationTarget: "github.com/NicoMoralesDev/optimusctx releases",
				Audience:          "Users who want the raw binary, need a fallback when package-manager metadata lags, or prefer explicit archive installs.",
				UpgradePath:       "Download a newer tagged archive, replace the binary on your PATH, and rerun the verification commands.",
				RollbackPath:      "Reinstall a prior tagged archive from GitHub Releases if an upgrade needs to be reversed.",
				SupportBoundary:   "Best-effort support covers download, verification, and explicit MCP setup; it does not extend to managed rollout tooling.",
			},
			{
				ID:                "homebrew",
				Name:              "Homebrew",
				UserFacingInstall: "Run `brew install niccrow/tap/optimusctx` on macOS or Linux once the formula is published.",
				PublicationTarget: "niccrow/homebrew-tap",
				Audience:          "macOS and Linux users who already rely on Homebrew for CLI distribution.",
				UpgradePath:       "Use `brew upgrade niccrow/tap/optimusctx`, then rerun `optimusctx version`, `optimusctx status`, and `optimusctx doctor`.",
				RollbackPath:      "If a package-manager rollback is needed, fall back to a prior tagged GitHub Release archive.",
				SupportBoundary:   "Support covers the published tap path and verification commands, not Homebrew-specific environment repair outside the documented flow.",
			},
			{
				ID:                "scoop",
				Name:              "Scoop",
				UserFacingInstall: "Run `scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git` and `scoop install niccrow/optimusctx` on Windows.",
				PublicationTarget: "niccrow/scoop-bucket",
				Audience:          "Windows users who already manage developer tooling through Scoop.",
				UpgradePath:       "Use `scoop update optimusctx`, then rerun `optimusctx version`, `optimusctx status`, and `optimusctx doctor`.",
				RollbackPath:      "If Scoop metadata is insufficient for a rollback, reinstall a prior tagged GitHub Release archive.",
				SupportBoundary:   "Support covers the named bucket, install command, and verification path, not broader Windows package-manager alternatives.",
			},
			{
				ID:                "npm",
				Name:              "npm",
				UserFacingInstall: "Run `npm install -g @niccrow/optimusctx` or `npx @niccrow/optimusctx version` to use the wrapper package over the tagged release binary.",
				PublicationTarget: "@niccrow/optimusctx",
				Audience:          "Users who already manage CLI tooling through npm or want an `npx` wrapper without changing the single-binary runtime contract.",
				UpgradePath:       "Use `npm install -g @niccrow/optimusctx@latest` or rerun `npx @niccrow/optimusctx version`, then rerun `optimusctx version`, `optimusctx status`, and `optimusctx doctor` if the binary stays installed locally.",
				RollbackPath:      "If the npm wrapper path regresses, reinstall a prior tagged archive directly from GitHub Releases or rerun the wrapper against the desired tagged version.",
				SupportBoundary:   "Support covers the published npm package and wrapper verification flow, not arbitrary Node environment repair or a JavaScript reimplementation path.",
			},
		},
		Upgrade: UpgradePolicy{
			VerificationCommands: []string{
				"optimusctx version",
				"optimusctx status",
				"optimusctx doctor",
			},
			ArchiveExpectation: "Archive users upgrade by replacing the binary manually and verifying the shipped command surface again.",
			PackageManagerRule: "Package-manager users upgrade through the channel-native command while GitHub Release archives remain the rollback fallback, including the npm wrapper package.",
		},
		Support: SupportPolicy{
			SupportedCommands: []string{
				"optimusctx version",
				"optimusctx init",
				"optimusctx status",
				"optimusctx doctor",
				"optimusctx init --client claude-desktop --write",
			},
			IssueTracking:      "Support is issue-driven through repository documentation and GitHub issues rather than a managed installer or helpdesk.",
			OperatorModel:      "Best-effort support assumes a user can rerun documented commands locally and report the exact failing step.",
			ConfigMutationRule: "`optimusctx init --client ...` is preview-first; host registration is only written when the operator opts into `--write`.",
		},
		DeferredScope: []DeferredScopeItem{
			{Name: "native Linux packages", Reason: "`.deb` and `.rpm` packaging is deferred until v2 expansion work."},
			{Name: "WinGet", Reason: "Windows distribution stays on Scoop plus GitHub Release archives for v1.2."},
			{Name: "Chocolatey", Reason: "Only one native Windows package-manager channel is supported in v1.2."},
			{Name: "artifact signing", Reason: "Signed artifacts are deferred to later distribution-hardening work."},
			{Name: "SBOM publication", Reason: "SBOM generation and publication remain out of scope for v1.2."},
		},
	}
}
