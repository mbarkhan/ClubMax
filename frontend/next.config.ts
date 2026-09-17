import type {NextConfig} from "next";

// Keep the development compiler isolated from `next build`. Running a
// production build while the dev server is active must not replace its chunks.
const internalAPI = (process.env.CLUBMAX_INTERNAL_API_URL || `http://127.0.0.1:${process.env.CLUBMAX_BACKEND_PORT || "8081"}`).replace(/\/+$/, "");

const nextConfig: NextConfig = {
  distDir: process.env.NODE_ENV === "development" ? ".next-dev" : ".next",
  output: "standalone",
  images: {unoptimized: true},
  async rewrites() {
    return [
      {source: "/v1/:path*", destination: `${internalAPI}/v1/:path*`},
      {source: "/media/:path*", destination: `${internalAPI}/media/:path*`},
    ];
  },
};

export default nextConfig;
