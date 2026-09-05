# sway-title-animator

<p align="center">
  <img src="docs/assets/logo.svg" alt="sway-title-animator logo" width="900">
</p>

Animated Unicode titlebars for Sway. The Linux-only Go program talks directly
to the Sway/i3 IPC socket and adds an application label, optional presentation
indicator, and animated art to window titles. It is especially useful with
Sway's `tabbed` and `stacked` layouts.

## Install

Build from source:

```sh
git clone https://github.com/marang/sway-title-animator.git
cd sway-title-animator
make install
```

This installs only `sway-title-animator` in `~/.local/bin`. Release packages
depend only on Sway. Sound-reactive presets additionally use the optional
`parec` command (provided by `libpulse` on Arch and `pulseaudio-utils` on
Debian/Ubuntu; PipeWire users can use `pipewire-pulse`).

On Arch Linux, install the AUR package with:

```sh
yay -S sway-title-animator
```

Linux release downloads, including `deb` and `rpm` packages for `amd64` and
`arm64`, are available from the
[GitHub Releases page](https://github.com/marang/sway-title-animator/releases).

## Sway setup

Add this line to your Sway configuration, or include
[`contrib/sway/45-title-animator.conf`](contrib/sway/45-title-animator.conf):

```conf
exec_always --no-startup-id /usr/bin/sway-title-animator --replace --fps 25
```

For a source install, replace `/usr/bin` with `$HOME/.local/bin`. Apply later
configuration changes with `swaymsg reload`.

## Presets and preview

List presets or preview every registered animation without a Sway connection:

```sh
sway-title-animator --list-presets
sway-title-animator --preview
```

The terminal preview puts each preset on a labeled line. Scroll with arrow
keys, `Page Up`/`Page Down`, or `Home`/`End`; press `q` or `Ctrl-C` to exit.
Select one preset with `--preset NAME`, or use `SWAY_TAB_ANIMATION=NAME`; the
configured rotation remains the default.

Each launch gets a fresh motion seed. A fixed seed remains deterministic for
tests. Sound variants use `<base>_sound`, preserve their base preset's motion,
fall back to that base when capture is unavailable, and are opt-in rather than
part of the default rotation. The sound-preset design record is in
[docs/sound-presets-plan.md](docs/sound-presets-plan.md).

## Configuration

Create a starting configuration with:

```sh
sway-title-animator --init-config
```

The default path is
`$XDG_CONFIG_HOME/sway-title-animator/config.toml` (or
`~/.config/sway-title-animator/config.toml`). See
[config.example.toml](config.example.toml) for all settings, including frame
rate, rotation, glyphs, audio sensitivity, icons, and presentation glyphs.

The animator decodes the stable `internal/titleindicator` Sway-mark protocol
when another tool supplies it. It does not create session state, run a session
daemon, or require a session manager.

## Moving to sway-session

Persistent work sessions now live in the separate
[sway-session repository](https://github.com/marang/sway-session). The split
is planned for `sway-title-animator` v0.10.0 and the matching first
`sway-session` release.

For an upgrade from a release at or below v0.9.3, the old animator package owns
both `sway-title-animator` and `sway-session`; the new packages own one binary
each. Do not remove the old package separately. Either upgrade the animator
package first and install `sway-session` immediately afterward without
restarting Sway, or build/download both new packages and install the pair in
one `pacman -U` invocation. Do not use `--overwrite`. Then replace the old
session lines in your Sway config with the template and instructions from the
new `sway-session` package, while keeping the animator line above. Package
metadata must not use a blanket `replaces` relationship that can remove the
animator while installing the session package.

## Development

Run the complete local/CI gate before handoff:

```sh
make verify
```

For animation changes, also inspect `--preview` at narrow and wide terminal
widths. Preserve exact requested output width, zero-width safety, fixed-seed
determinism, and the all-pairs visual-similarity guard.
