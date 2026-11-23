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

          myapp = pkgs.buildGo126Module {
          pname = appName;
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-2BtLHbiYVg+N70eQ22ynuf0+Z/6Q9Lqy/2e5VcANs4Y=";
          doCheck = false;

          nativeBuildInputs = [ pkgs.templ pkgs.tailwindcss_4 ];

          preBuild = ''
            rm -f internal/templates/*_templ.go
            templ generate
            tailwindcss -i ./static/css/input.css -o ./static/css/style.min.css --minify
          '';

          subPackages = [ "cmd" "cmd/agent" ];
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
              adminEmail = mkOption {
                type = types.str;
                default = "";
                description = "Login email of the admin user, sees the /admin page.";
              };
              agentToken = mkOption {
                type = types.str;
                default = "";
                description = "Shared secret agents must present to pair. Empty disables auto-registration.";
              };
              sessionSecret = mkOption {
                type = types.str;
                default = "";
                description = "Secret used to sign session cookies. Empty means sessions die on restart.";
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
                    "SERVERPORT=${toString cfg.port}"
                    "DATABASE_URL=sqlite:${cfg.dataDir}/goth.db"
                    "STATIC_DIR=${pkg}/share/goth/static"
                    "ADMIN_EMAIL=${cfg.adminEmail}"
                    "AGENT_TOKEN=${cfg.agentToken}"
                    "SESSION_SECRET=${cfg.sessionSecret}"
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

        nixosModules.agent = { config, lib, pkgs, ... }:
          with lib;
          let
            cfg = config.services.gothAgent;
            pkg = self.packages.${pkgs.system}.default;
          in {
            options.services.gothAgent = {
              enable = mkEnableOption "goth agent";
              token = mkOption {
                type = types.str;
                default = "";
                description = "Shared AGENT_TOKEN, must match the server's AGENT_TOKEN to pair.";
              };
              server = mkOption {
                type = types.str;
                default = "http://localhost:4000";
                description = "Base URL of the web app the agents phone home to.";
              };
              agents = mkOption {
                type = types.listOf (types.submodule {
                  options = {
                    name = mkOption { type = types.str; };
                    port = mkOption { type = types.port; };
                    version = mkOption { type = types.str; default = "1.0.0"; };
                    os = mkOption { type = types.str; default = "NixOS 25.05"; };
                    heartbeatSecs = mkOption { type = types.int; default = 30; };
                  };
                });
                default = [ ];
                example = [
                  { name = "lab-pc-1"; port = 4101; }
                  { name = "lab-pc-2"; port = 4102; }
                ];
                description = "Agent instances to run, one systemd service each. They self-register on the server.";
              };
            };

            config = mkIf cfg.enable {
              systemd.services = listToAttrs (map
                (a: {
                  name = "goth-agent-${a.name}";
                  value = {
                    description = "Goth agent ${a.name}";
                    after = [ "network.target" ];
                    wantedBy = [ "multi-user.target" ];
                    serviceConfig = {
                      Type = "simple";
                      DynamicUser = true;
                      ExecStart = "${pkg}/bin/agent";
                      Restart = "always";
                      Environment = [
                        "AGENT_NAME=${a.name}"
                        "AGENT_VERSION=${a.version}"
                        "AGENT_OS=${a.os}"
                        "AGENT_HEARTBEAT_SECS=${toString a.heartbeatSecs}"
                        "AGENT_TOKEN=${cfg.token}"
                        "AGENT_SERVER=${cfg.server}"
                        "AGENT_URL=http://localhost:${toString a.port}"
                        "PORT=:${toString a.port}"
                      ];
                    };
                  };
                })
                cfg.agents);
            };
          };
      };
}

