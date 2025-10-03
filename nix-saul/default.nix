{ pkgs ? import <nixpkgs> {} }:

pkgs.buildGoModule {
  pname = "better-curl-saul";
  version = "v0.3.0";

  src = pkgs.fetchFromGitHub {
    owner = "DeprecatedLuar";
    repo = "better-curl-saul";
    rev = "main";
    sha256 = "sha256-KIJndQxICxDB7w6snQeserCVTVC8u/ueFfwo5L8souQ=";
  };

  vendorHash = "sha256-h/W5e64XQmfDgW6JPgxOJ1Jw8B18SsaX31nDvPTAQHI=";

  subPackages = [ "cmd" ];

  postInstall = ''
    mv $out/bin/cmd $out/bin/saul
  '';

  meta = with pkgs.lib; {
    description = "Did you know you have rights? The FOSS says you do... Better Curl Saul is my homemade CLI 'http client' to make api reusability simple";
    license = licenses.mit;
    maintainers = [ "DeprecatedLuar" ];
    homepage = "https://github.com/DeprecatedLuar/better-curl-saul";
    platforms = platforms.linux ++ platforms.darwin ++ platforms.windows;
  };
}
