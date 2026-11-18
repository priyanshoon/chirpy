{
  description = "chirpy - go project";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    systems.url = "github:nix-systems/default";
    process-compose-flake.url = "github:Platonic-Systems/process-compose-flake";
    services-flake.url = "github:juspay/services-flake";
  };
  outputs = inputs:
    inputs.flake-parts.lib.mkFlake {inherit inputs;} {
      systems = import inputs.systems;
      imports = [
        inputs.process-compose-flake.flakeModule
      ];
      perSystem = {
        self',
        pkgs,
        config,
        ...
      }: {
        process-compose."chirpy" = {config, ...}: let
          dbName = "chirpy";
        in {
          imports = [
            inputs.services-flake.processComposeModules.default
          ];

          services.postgres."pg1" = {
            enable = true;
            # initialDatabase = [
            #   {
            #     name = dbName;
            #   }
            # ];
          };

          settings.processes.pgweb = let
            pgcfg = config.services.postgres.pg1;
          in {
            environment.PGWEB_DATABASE_URL = pgcfg.connectionURI {inherit dbName;};
            command = pkgs.pgweb;
            depends_on."pg1".condition = "process_healthy";
          };
        };
        packages.default = self'.packages.chirpy;

        devShells.default = pkgs.mkShell {
          inputsFrom = [
            config.process-compose."chirpy".services.outputs.devShell
          ];
          packages = [
            pkgs.goose
            pkgs.sqlc
            self'.packages.chirpy
          ];
        };
      };
    };
}
