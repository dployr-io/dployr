// vite.config.ts
import { defineConfig } from "file:///C:/Workspace/dployr/client/frontend/node_modules/.pnpm/vite@5.4.19/node_modules/vite/dist/node/index.js";
import { svelte } from "file:///C:/Workspace/dployr/client/frontend/node_modules/.pnpm/@sveltejs+vite-plugin-svelte@3.1.2_svelte@4.2.20_vite@5.4.19/node_modules/@sveltejs/vite-plugin-svelte/src/index.js";
import Icons from "file:///C:/Workspace/dployr/client/frontend/node_modules/.pnpm/unplugin-icons@22.1.0_svelte@4.2.20/node_modules/unplugin-icons/dist/vite.js";
var vite_config_default = defineConfig({
  plugins: [
    svelte(),
    Icons({
      compiler: "svelte",
      autoInstall: true
    })
  ],
  css: {
    postcss: "./postcss.config.js"
  }
});
export {
  vite_config_default as default
};
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsidml0ZS5jb25maWcudHMiXSwKICAic291cmNlc0NvbnRlbnQiOiBbImNvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9kaXJuYW1lID0gXCJDOlxcXFxXb3Jrc3BhY2VcXFxcZHBsb3lyXFxcXGNsaWVudFxcXFxmcm9udGVuZFwiO2NvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9maWxlbmFtZSA9IFwiQzpcXFxcV29ya3NwYWNlXFxcXGRwbG95clxcXFxjbGllbnRcXFxcZnJvbnRlbmRcXFxcdml0ZS5jb25maWcudHNcIjtjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfaW1wb3J0X21ldGFfdXJsID0gXCJmaWxlOi8vL0M6L1dvcmtzcGFjZS9kcGxveXIvY2xpZW50L2Zyb250ZW5kL3ZpdGUuY29uZmlnLnRzXCI7aW1wb3J0IHsgZGVmaW5lQ29uZmlnIH0gZnJvbSBcInZpdGVcIjtcclxuaW1wb3J0IHsgc3ZlbHRlIH0gZnJvbSBcIkBzdmVsdGVqcy92aXRlLXBsdWdpbi1zdmVsdGVcIjtcclxuaW1wb3J0IEljb25zIGZyb20gXCJ1bnBsdWdpbi1pY29ucy92aXRlXCI7XHJcblxyXG4vLyBodHRwczovL3ZpdGVqcy5kZXYvY29uZmlnL1xyXG5leHBvcnQgZGVmYXVsdCBkZWZpbmVDb25maWcoe1xyXG4gIHBsdWdpbnM6IFtcclxuICAgIHN2ZWx0ZSgpLFxyXG4gICAgSWNvbnMoe1xyXG4gICAgICBjb21waWxlcjogXCJzdmVsdGVcIixcclxuICAgICAgYXV0b0luc3RhbGw6IHRydWUsXHJcbiAgICB9KSxcclxuICBdLFxyXG4gIGNzczoge1xyXG4gICAgcG9zdGNzczogXCIuL3Bvc3Rjc3MuY29uZmlnLmpzXCIsXHJcbiAgfSxcclxufSk7XHJcbiJdLAogICJtYXBwaW5ncyI6ICI7QUFBcVMsU0FBUyxvQkFBb0I7QUFDbFUsU0FBUyxjQUFjO0FBQ3ZCLE9BQU8sV0FBVztBQUdsQixJQUFPLHNCQUFRLGFBQWE7QUFBQSxFQUMxQixTQUFTO0FBQUEsSUFDUCxPQUFPO0FBQUEsSUFDUCxNQUFNO0FBQUEsTUFDSixVQUFVO0FBQUEsTUFDVixhQUFhO0FBQUEsSUFDZixDQUFDO0FBQUEsRUFDSDtBQUFBLEVBQ0EsS0FBSztBQUFBLElBQ0gsU0FBUztBQUFBLEVBQ1g7QUFDRixDQUFDOyIsCiAgIm5hbWVzIjogW10KfQo=
