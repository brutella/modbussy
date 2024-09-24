# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Modbussy < Formula
  desc "Access modbus networks from the command line"
  homepage "https://hochgatterer.me"
  version "1.0.0-beta1"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/brutella/modbussy/releases/download/1.0.0-beta1/modbussy_Darwin_x86_64.tar.gz", using: CurlDownloadStrategy,
        headers: [
          "Accept: application/octet-stream",
          "Authorization: bearer #{ENV["HOMEBREW_GITHUB_API_TOKEN"]}"
        ]
      sha256 "a243eaf72ab3fbd6952298195cfcb5609b0ba56e8f4755751a06ee96fe30978a"

      def install
        bin.install "modbussy"
      end
    end
    on_arm do
      url "https://github.com/brutella/modbussy/releases/download/1.0.0-beta1/modbussy_Darwin_arm64.tar.gz", using: CurlDownloadStrategy,
        headers: [
          "Accept: application/octet-stream",
          "Authorization: bearer #{ENV["HOMEBREW_GITHUB_API_TOKEN"]}"
        ]
      sha256 "9eda74e5b9a48ebc101f12324def8bbd6e64dcdd69f846dac57608e522c5cabd"

      def install
        bin.install "modbussy"
      end
    end
  end

  on_linux do
    on_intel do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/brutella/modbussy/releases/download/1.0.0-beta1/modbussy_Linux_x86_64.tar.gz", using: CurlDownloadStrategy,
          headers: [
            "Accept: application/octet-stream",
            "Authorization: bearer #{ENV["HOMEBREW_GITHUB_API_TOKEN"]}"
          ]
        sha256 "72cd4bdfec23b300dcc15160f9611e0c7ade264df64141546c45034dcc48f307"

        def install
          bin.install "modbussy"
        end
      end
    end
    on_arm do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/brutella/modbussy/releases/download/1.0.0-beta1/modbussy_Linux_arm64.tar.gz", using: CurlDownloadStrategy,
          headers: [
            "Accept: application/octet-stream",
            "Authorization: bearer #{ENV["HOMEBREW_GITHUB_API_TOKEN"]}"
          ]
        sha256 "ea89b25d8055e1393aa763dcf81c2e8daef1ee7d1808f3a14b14c0d7b85f903f"

        def install
          bin.install "modbussy"
        end
      end
    end
  end
end