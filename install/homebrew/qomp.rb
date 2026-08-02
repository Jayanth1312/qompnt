class Qomp < Formula
  desc "Install qompnt themes and HTML components into a project"
  homepage "https://qompnt.vercel.app"
  version "1.0.6"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.6/qomp_darwin_arm64.tar.gz"
      sha256 "8d11df842db7b3bf0384adca559591ed806eda02918e5a25a76024f477adc713"
    end
    on_intel do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.6/qomp_darwin_amd64.tar.gz"
      sha256 "ee73b7ad9ffb237dee379e899da713b2f00d950cfd787a505f580b2c65a6c972"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.6/qomp_linux_arm64.tar.gz"
      sha256 "1e0fb214783b88ffef9781ce99a95d33fb6b4deef9e1af28e313115922fc2dd5"
    end
    on_intel do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.6/qomp_linux_amd64.tar.gz"
      sha256 "e27c9d92f89854439bceb25534c2c479ec9fdce774a168b0a3f540d44deccdd3"
    end
  end

  def install
    bin.install "qomp"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/qomp --version")
  end
end
