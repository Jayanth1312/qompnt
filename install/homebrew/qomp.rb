class Qomp < Formula
  desc "Install qompnt themes and HTML components into a project"
  homepage "https://qompnt.vercel.app"
  version "1.0.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.0/qomp_darwin_arm64.tar.gz"
      sha256 "9ce70dd386832bc15c7910878c6adbe2851ec2b4305982695fb0523de6212b89"
    end
    on_intel do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.0/qomp_darwin_amd64.tar.gz"
      sha256 "081edcdd19b02d996b8af01212a992cef6bf68cda581eb60e235426565feabef"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.0/qomp_linux_arm64.tar.gz"
      sha256 "0b9a4a024e9193435b49dad62bb4191bf660908c03b90c25e1beea3580eb8c61"
    end
    on_intel do
      url "https://github.com/Jayanth1312/qompnt/releases/download/v1.0.0/qomp_linux_amd64.tar.gz"
      sha256 "faa5a85635ca1abc731582b0a447941e242a77d39f91e6abb6d65a0043f7145b"
    end
  end

  def install
    bin.install "qomp"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/qomp --version")
  end
end
