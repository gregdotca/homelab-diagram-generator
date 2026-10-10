# Changelog

## 2026-10-09

### Added

- Added an MIT `LICENSE` and a License section at the end of the
  README. Without a license, nobody else was allowed to use, change,
  or share the code.
- Added `ACR.md` with an [AI Code Rating](https://aicoderating.com)
  of **ACR B4c**, plus a badge at the top of the README, so people
  can see how much of the code AI wrote and how carefully it was
  checked.

### Changed

- Renamed the GitHub repo from `gregdotca/hdg` to
  `gregdotca/homelab-diagram-generator`, so the name says what the
  tool does and people searching GitHub can find it.
- Changed the Go module path in `go.mod` to match the new repo, and
  updated the internal imports in `main.go`,
  `internal/layout/layout.go`, and `internal/render/svg.go`. Without
  this, `go install` from the new URL fails with a module path
  mismatch.
- Renamed the binary from `homelab` to `hdg`. `build.sh`, the README,
  and the binary in the repo all used different names before.
- Fixed the example command in `build.sh`. It pointed to
  `example/homelab.yaml`, which doesn't exist, and now uses
  `sample.yaml`.
- Rewrote the README. It was a few lines, and now covers what the
  tool does, features, installation, usage, every config key, and
  the icon list, so people who find the repo can tell what it is
  and search engines have text to match.
- Updated the header comment in `sample.yaml` to use the new name
  and the `hdg` binary.
- Ran `gofmt` on `internal/config/config.go`,
  `internal/icons/icons.go`, `internal/layout/layout.go`, and
  `internal/render/svg.go`. The struct and constant columns were
  over-padded and there were stray blank lines, so `gofmt -l` flagged
  all four files. Spacing only; the code behaves the same.
