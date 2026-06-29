/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  basePath: '/ai-daily',
  assetPrefix: '/ai-daily',
  images: {
    unoptimized: true,
  },
  trailingSlash: true,
  // 使用 SWC 压缩（Next.js 14 默认）
  swcMinify: true,
};

module.exports = nextConfig;
