# Releasing sway-title-animator

Run `make verify`, review the complete diff, and use the release workflow only
from a tag whose commit is on `main`. GoReleaser produces Linux archives and
packages containing only `sway-title-animator`, its configuration example, and
the animator Sway template. Runtime package dependency is Sway only.

The checked-in `PKGBUILD` source metadata and `.SRCINFO` remain pinned to the
last released v0.9.3 archive and its verified checksum, while the package body
is the template for the next animator-only release. The recipe rejects legacy
combined source to prevent packaging stale session startup references. Do not change that checksum
to a placeholder or `SKIP`; the AUR workflow replaces version, release number,
and checksum from the tagged archive and opens a metadata-sync PR after a
verified build.

## v0.10.0 paired cutover

v0.10.0 removes `sway-session` from this package. Coordinate its release with
the separate [sway-session](https://github.com/marang/sway-session) package.
Before publishing, verify that users upgrading from v0.9.3 or earlier can
upgrade the animator first then install the session package without restarting
Sway, or install both built files together with `pacman -U`; do not claim an
unverified helper update is atomic and do not require `--overwrite`. Neither
package may use a blanket `replaces` declaration that removes the other
package. Publish the migration guide in both release notes and check the final
artifacts for exactly one owner of each binary and template.
