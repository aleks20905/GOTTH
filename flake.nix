{
  description = "Go + Templ + Tailwind development server";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    templ.url = "github:a-h/templ";
  };

  outputs = { self, nixpkgs, flake-utils, templ }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        templPkg = templ.packages.${system}.templ;
        
        # Your app name from Makefile
        appName = "yourapp"; # Change this to your actual app name
        
        # Tailwind CSS standalone binary
        tailwindcss = pkgs.stdenv.mkDerivation rec {
          pname = "tailwindcss";
          version = "4.1.18"; # Update to latest if needed
          
          src = pkgs.fetchurl {
            url = "https://github.com/tailwindlabs/tailwindcss/releases/download/v${version}/tailwindcss-linux-x64";
            sha256 = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # Will fix this
          };
          
          dontUnpack = true;
          dontBuild = true;
          
          installPhase = ''
            mkdir -p $out/bin
            cp $src $out/bin/tailwindcss
            chmod +x $out/bin/tailwindcss
          '';
        };

        # Air for live reload
        air = pkgs.buildGoModule rec {
          pname = "air";
          version = "1.49.0";
          
          src = pkgs.fetchFromGitHub {
            owner = "air-verse";
            repo = "air";
            rev = "v${version}";
            sha256 = "sha256-BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB="; # Will fix this
          };
          
          vendorHash = "sha256-CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC=";
        };

      in {
        # Development shell with all tools
        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.gopls
            pkgs.gotools
            templPkg
            tailwindcss
            air
            pkgs.sqlite
            pkgs.goose # Optional: for DB migrations
          ];
          
          shellHook = ''
            echo "🚀 Go + Templ + Tailwind dev environment loaded"
            echo "Available commands:"
            echo "  make tailwind-watch  - Watch CSS changes"
            echo "  make templ-watch     - Watch template changes"  
            echo "  make dev             - Start air live reload"
            echo "  air                  - Direct air command"
          '';
        };

        # Package for building
        packages.default = pkgs.buildGoModule {
          pname = appName;
          version = "0.1.0";
          src = ./.;
          
          vendorHash = pkgs.lib.fakeHash; # Run once to get real hash
          
          nativeBuildInputs = [ templPkg tailwindcss ];
          
          preBuild = ''
            # Generate templ files
            templ generate
            
            # Build CSS
            tailwindcss -i ./static/css/input.css -o ./static/css/style.min.css --minify
          '';
          
          ldflags = [
            "-X main.Environment=production"
          ];
          
          # Include static files
          postInstall = ''
            mkdir -p $out/share/${appName}
            cp -r ./static $out/share/${appName}/
          '';
        };

        # NixOS module for server deployment
        nixosModules.default = { config, lib, pkgs, ... }:
          with lib;
          let
            cfg = config.services.${appName};
            pkg = self.packages.${system}.default;
          in {
            options.services.${appName} = {
              enable = mkEnableOption "Enable ${appName} service";
              
              port = mkOption {
                type = types.port;
                default = 8080;
                description = "Port to listen on";
              };
              
              dataDir = mkOption {
                type = types.path;
                default = "/var/lib/${appName}";
                description = "Directory for SQLite database";
              };
              
              openFirewall = mkOption {
                type = types.bool;
                default = true;
                description = "Open firewall port";
              };
            };
            
            config = mkIf cfg.enable {
              # Create user and group
              users.users.${appName} = {
                isSystemUser = true;
                group = appName;
                home = cfg.dataDir;
                createHome = true;
              };
              
              users.groups.${appName} = {};
              
              # Systemd service
              systemd.services.${appName} = {
                description = "${appName} Go web server";
                after = [ "network.target" ];
                wantedBy = [ "multi-user.target" ];
                
                serviceConfig = {
                  Type = "simple";
                  User = appName;
                  Group = appName;
                  WorkingDirectory = cfg.dataDir;
                  
                  ExecStart = "${pkg}/bin/${appName}";
                  
                  # Environment variables
                  Environment = [
                    "PORT=${toString cfg.port}"
                    "ENVIRONMENT=production"
                    "DATA_DIR=${cfg.dataDir}"
                  ];
                  
                  # Security hardening (relaxed for dev/friend-preview)
                  NoNewPrivileges = true;
                  ProtectSystem = "strict";
                  ProtectHome = true;
                  ReadWritePaths = [ cfg.dataDir ];
                };
              };
              
              # Firewall
              networking.firewall.allowedTCPPorts = mkIf cfg.openFirewall [ cfg.port ];
              
              # Ensure data directory exists with correct permissions
              systemd.tmpfiles.rules = [
                "d ${cfg.dataDir} 0750 ${appName} ${appName} -"
              ];
            };
          };
      });
}