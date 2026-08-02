class Qomp < Formula
  desc "Install qompnt themes and HTML components into a project"
  homepage "https://qompnt.vercel.app"
  version "1.0.2"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.2/qomp_darwin_arm64.tar.gz"
      sha256 "426b2d1e17a49b3d9ca631a61a8f1ebf9a3cb3fe77e1aea514134d4755b7cd4d"
    end
    on_intel do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.2/qomp_darwin_amd64.tar.gz"
      sha256 "79de1f6f90ff1275faba3256ef7e7b2e19c46f2ce7c1082bd916a342987c85a5"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.2/qomp_linux_arm64.tar.gz"
      sha256 "769e1808a9823ce69dc803ac0de400ddc1af2905599081006cfa5458344f4e6a"
    end
    on_intel do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.2/qomp_linux_amd64.tar.gz"
      sha256 "b63436da1cf859d2ab6b71a08fc19e5e587f4bf61b16e0b25b73f834ade86852"
    end
  end

  def install
    bin.install "qomp"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/qomp --version")
  end
end
