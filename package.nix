{
  lib,
  buildGoModule,
}:
buildGoModule {
  pname = "deepstate";
  version = "0.0.0";

  src = lib.cleanSource ./.;

  vendorHash = "sha256-Ym9+nVIv/xMeqBRtn/WJ2XhHglL3n6q11pSqHesU6ew=";

  meta = {
    description = "Scan for leftovers";
    license = lib.licenses.gpl3Only;
    mainProgram = "deepstate";
  };
}
