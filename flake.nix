{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable-small";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ self.overlay ];
          };
        in
        {
          packages.default = pkgs.terrasops;
        }
      ) //
    {
      overlay = final: prev: {
        terrasops = final.buildGoModule {
          pname = "terrasops";
          version = "unstable-${final.lib.substring 0 8 self.lastModifiedDate}";
          src = self;
          vendorHash = "sha256-Ttz1nkJGOZEyjxYsgeqzjKdao5Ju2MTlAwerXTD/aw4=";
        };
      };
    };
}
