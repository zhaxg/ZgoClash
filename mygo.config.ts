export default {
  name: "ZgoClash",
  identifier: "app.zgo.clash",
  version: "1.0.1",
  description: "A compact mihomo client for Windows",
  out: "bin",
  // MyGo embeds the exe icon from a square PNG only, so the artwork in
  // resources/app.ico (a 64 px BMP-only .ico) is kept as icon.png beside it.
  // Both files ship in resources/; the .ico is never read by the build.
  icon: "resources/icon.png",
  windows: {
    displayName: "ZgoClash",
  },
};
