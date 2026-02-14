{
  description = "Production Go + Templ app";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        appName = "goth";

        myapp = pkgs.buildGoModule {
          pname = appName;
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-NEDkSYvYrIM05TJ2YGTyiTcrubCeqSqKbLh5bTswIJA=";
          doCheck = false;

          nativeBuildInputs = [ pkgs.templ pkgs.tailwindcss_4 ];

          preBuild = ''
            rm -f internal/templates/*_templ.go
            templ generate
            tailwindcss -i ./static/css/input.css -o ./static/css/style.min.css --minify
          '';

          subPackages = [ "cmd" ];
          ldflags = [ "-s" "-w" "-X main.Environment=production" ];

          postInstall = ''
            mkdir -p $out/share/goth
            cp -r ./static $out/share/goth/
            mv $out/bin/cmd $out/bin/${appName}
          '';
        };

      in {
        packages.default = myapp;

        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.gopls
            pkgs.templ
            pkgs.tailwindcss_4
            pkgs.air
            pkgs.sqlite
          ];
        };
      }) // {
        nixosModules.default = { config, lib, pkgs, ... }:
          with lib;
          let
            cfg = config.services.goth;
            pkg = self.packages.${pkgs.system}.default;
          in {
            options.services.goth = {
              enable = mkEnableOption "goth service";
              port = mkOption {
                type = types.port;
                default = 4000;
              };
              dataDir = mkOption {
                type = types.path;
                default = "/var/lib/goth";
              };
              openFirewall = mkOption {
                type = types.bool;
                default = false;
              };
            };

            config = mkIf cfg.enable {
              users.users.goth = {
                isSystemUser = true;
                group = "goth";
                home = cfg.dataDir;
                createHome = true;
              };
              users.groups.goth = { };
              systemd.tmpfiles.rules = [ "d ${cfg.dataDir} 0750 goth goth -" ];

              systemd.services.goth = {
                description = "Goth web app";
                after = [ "network.target" ];
                wantedBy = [ "multi-user.target" ];
                serviceConfig = {
                  Type = "simple";
                  User = "goth";
                  Group = "goth";
                  ExecStart = "${pkg}/bin/goth";
                  WorkingDirectory = cfg.dataDir;
                  Restart = "always";

                  #environt stuff !!!
                  Environment = [
                    "PORT=${toString cfg.port}"
                    "DATABASE_URL=sqlite:${cfg.dataDir}/goth.db"
                    "STATIC_DIR=${pkg}/share/goth/static"

                  ];
                  #environt stuff !!!

                  NoNewPrivileges = true;
                  PrivateTmp = true;
                  ProtectSystem = "strict";
                  ProtectHome = true;
                  ReadWritePaths = [ cfg.dataDir ];
                };
              };

              networking.firewall.allowedTCPPorts =
                mkIf cfg.openFirewall [ cfg.port ];

            };
          };
      };
}

