# Reference formula — GoReleaser overwrites this file in the repo on each release.
# To install from a local build, use: make install
# To install from Homebrew:
#   brew tap lovelyJason/openskills https://github.com/lovelyJason/openskills
#   brew install openskills

class Openskills < Formula
  desc "AI editor extension manager — plugins, skills, marketplaces for Codex, Claude, Cursor"
  homepage "https://github.com/lovelyJason/openskills"
  license "MIT"
  version "0.1.0"

  on_macos do
    on_arm do
      url "https://github.com/lovelyJason/openskills/releases/download/v#{version}/openskills_#{version}_darwin_arm64.tar.gz"
      # sha256 is auto-filled by GoReleaser
    end
    on_intel do
      url "https://github.com/lovelyJason/openskills/releases/download/v#{version}/openskills_#{version}_darwin_amd64.tar.gz"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/lovelyJason/openskills/releases/download/v#{version}/openskills_#{version}_linux_arm64.tar.gz"
    end
    on_intel do
      url "https://github.com/lovelyJason/openskills/releases/download/v#{version}/openskills_#{version}_linux_amd64.tar.gz"
    end
  end

  def install
    bin.install "openskills"
  end

  def caveats
    <<~EOS
      🛠️  openskills installed successfully!

      Get started:
        openskills marketplace add <git-url>    # add a marketplace
        openskills plugin list                   # browse plugins
        openskills skill list                    # browse skills
        openskills status                        # check system

      Docs: https://github.com/lovelyJason/openskills
    EOS
  end

  test do
    system "#{bin}/openskills", "--version"
  end
end
