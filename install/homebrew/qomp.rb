class Qomp < Formula
  desc "Install qompnt themes and HTML components into a project"
  homepage "https://qompnt.vercel.app"
  version "1.0.5"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.5/qomp_darwin_arm64.tar.gz"
      sha256 "c61788accad22ed8bb13195cf04687adeb27e418ad67acc4a5acae071cca9835"
    end
    on_intel do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.5/qomp_darwin_amd64.tar.gz"
      sha256 "39e490b5e6eac6b4a1e289c56a7416a5092ac57489e6af0847d342dce948ee12"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.5/qomp_linux_arm64.tar.gz"
      sha256 "9a1e1ea3405cb3860a18b36d3a42f276fb94e5a7d2414dc589de06c2f917de2b"
    end
    on_intel do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.5/qomp_linux_amd64.tar.gz"
      sha256 "5b8825b4b3100c37198d920047ed9a35a5a13cd236830378e3911db0d08379f9"
    end
  end

  def install
    bin.install "qomp"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/qomp --version")
  end
end
