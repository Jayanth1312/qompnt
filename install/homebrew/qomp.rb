class Qomp < Formula
  desc "Install qompnt themes and HTML components into a project"
  homepage "https://qompnt.vercel.app"
  version "1.0.3"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.3/qomp_darwin_arm64.tar.gz"
      sha256 "81f6887add529f68f86667619a0509ce00d15e3da566c4eb710991cf08b47e29"
    end
    on_intel do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.3/qomp_darwin_amd64.tar.gz"
      sha256 "9f8e665fc8682aad0b6caac39a2a4b5ac205c84fee42de976ff8c254ea42b69b"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.3/qomp_linux_arm64.tar.gz"
      sha256 "d29811160f0146193052d5cf7dcd5ef3aa4def8f52b3e4b074f122df0f7c6423"
    end
    on_intel do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.3/qomp_linux_amd64.tar.gz"
      sha256 "41564ab071ec2cd9875fd5fc1602067e31b226032776320819aa554653d561d4"
    end
  end

  def install
    bin.install "qomp"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/qomp --version")
  end
end
