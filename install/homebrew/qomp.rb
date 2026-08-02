class Qomp < Formula
  desc "Install qompnt themes and HTML components into a project"
  homepage "https://qompnt.vercel.app"
  version "1.0.4"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.4/qomp_darwin_arm64.tar.gz"
      sha256 "7c4ad2ee93ea03d0cf151482d90a39a5c135814722e849759c612498ec385216"
    end
    on_intel do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.4/qomp_darwin_amd64.tar.gz"
      sha256 "c937a972a3f4ee99467b9ffbcc04806d6ea14e233b6a5745e2b453573ba0e381"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.4/qomp_linux_arm64.tar.gz"
      sha256 "1df3b93d66070eaceb017b7a98001a249c3ce5ece1e429e45c902a84536d7eb5"
    end
    on_intel do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.4/qomp_linux_amd64.tar.gz"
      sha256 "ccc4a0ccf241c14f59da8d7c6093e970c8832b89571c58a7575d303d435a1d30"
    end
  end

  def install
    bin.install "qomp"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/qomp --version")
  end
end
