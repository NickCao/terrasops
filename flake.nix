{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable-small";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let pkgs = import nixpkgs {
          inherit system;
          overlays = [ self.overlay ];
        }; in
        rec {
          packages.default = pkgs.terrasops;
        }
      ) //
    {
      overlay = final: prev: {
        terrasops = final.buildGoModule {
          pname = "terrasops";
          version = "unstable-${final.lib.substring 0 8 self.lastModifiedDate}";
          src = self;
          vendorSha256 = "sha256-SSvhBPYMuwjaGH++2oOO3fYUXV5T+yz5nqD0y4zep3g=";
        };
      };
    };
}
