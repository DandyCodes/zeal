{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = inputs:
    inputs.flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import inputs.nixpkgs {
          inherit system;
          #config.allowUnfree = true;
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            #BUILD
            go
            nodejs

            #DEV
            just
            air
            nixd
            nixpkgs-fmt
            nixfmt
          ];
        };
      in
      { inherit devShells; });
}
