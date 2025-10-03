{
  description = "Better Curl Saul - FOSS HTTP client";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in {
        packages.default = pkgs.callPackage ./default.nix {};
        devShell = pkgs.callPackage ./shell.nix {};
        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/saul";
        };
      }
    );
}
