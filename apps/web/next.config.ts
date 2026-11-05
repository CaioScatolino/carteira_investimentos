import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    return [
      { source: "/carteira", destination: "/" },
      { source: "/proventos", destination: "/" },
      { source: "/calendario", destination: "/" },
    ];
  },
};

export default nextConfig;
