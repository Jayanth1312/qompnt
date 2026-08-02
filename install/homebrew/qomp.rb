class Qomp < Formula
  desc "Install qompnt themes and HTML components into a project"
  homepage "https://qompnt.vercel.app"
  version "1.0.1"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.1/qomp_darwin_arm64.tar.gz"
      sha256 "74a61a2553e706a4497fb0cb2fe19261c73ebbc663e2f222fc9ac061e917e16d"
    end
    on_intel do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.1/qomp_darwin_amd64.tar.gz"
      sha256 "6d6f3f6512b86d23ad58f21928554050d9e0824fd03329f637b477c506814211"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.1/qomp_linux_arm64.tar.gz"
      sha256 "bab124d1fb8034a64d6fc831613fc1f143a6fc336503aacc5b2d130e9b423056"
    end
    on_intel do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.1/qomp_linux_amd64.tar.gz"
      sha256 "a1beffc628012030ae829f328b0877b429b56411caf70ccf6ab43bc22299029f"
    end
  end

  def install
    bin.install "qomp"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/qomp --version")
  end
end
