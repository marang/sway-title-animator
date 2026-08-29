pkgname=sway-title-animator
pkgver=0.1.0
pkgrel=1
pkgdesc="Animated Unicode titlebars for Sway"
arch=('x86_64' 'aarch64')
url="https://github.com/marang/sway-title-animator"
license=('MIT')
depends=('sway')
optdepends=(
  'libpulse: sound-reactive animation presets via parec'
  'alacritty: persistent work-session windows'
  'herdr: persistent terminal panes, history, and agent sessions'
  'apparmor: secure Codex resume boundary'
)
makedepends=('go>=1.26')
source=("sway-title-animator-$pkgver.tar.gz::$url/archive/refs/tags/v$pkgver.tar.gz")
sha256sums=('SKIP')

build() {
  cd "sway-title-animator-$pkgver"
  CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o sway-title-animator ./cmd/sway-title-animator
  CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o sway-session ./cmd/sway-session
}

package() {
  cd "sway-title-animator-$pkgver"
  install -Dm755 sway-title-animator "$pkgdir/usr/bin/sway-title-animator"
  install -Dm755 sway-session "$pkgdir/usr/bin/sway-session"
  install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
  install -Dm644 README.md "$pkgdir/usr/share/doc/$pkgname/README.md"
  install -Dm644 config.example.toml "$pkgdir/usr/share/doc/$pkgname/config.example.toml"
  install -Dm644 contrib/sway/45-title-animator.conf "$pkgdir/usr/share/doc/$pkgname/45-title-animator.conf"
  install -Dm644 contrib/herdr/config.toml "$pkgdir/usr/share/doc/$pkgname/contrib/herdr/config.toml"
  install -Dm644 contrib/codex/hooks-system.json "$pkgdir/usr/share/doc/$pkgname/contrib/codex/hooks.json"
  install -Dm644 contrib/apparmor/codex-home-guard "$pkgdir/usr/share/doc/$pkgname/contrib/apparmor/codex-home-guard"
  install -Dm755 scripts/verify-codex-boundary.sh "$pkgdir/usr/share/doc/$pkgname/scripts/verify-codex-boundary.sh"
}
