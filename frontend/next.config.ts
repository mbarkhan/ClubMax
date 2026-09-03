import type {NextConfig} from "next";

// Keep the development compiler isolated from `next build`. Running a
// production build while the dev server is active must not replace its chunks.
const nextConfig: NextConfig = {
  distDir: process.env.NODE_ENV === "development" ? ".next-dev" : ".next",
  output: "standalone",
  images: {unoptimized: true},
};

export default nextConfig;
