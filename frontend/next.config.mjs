import createNextIntlPlugin from "next-intl/plugin";

const withNextIntl = createNextIntlPlugin("./src/shared/i18n/request.ts");

const configuredRemotePatterns = [
  {
    protocol: "http",
    hostname: "localhost",
    port: "9000",
    pathname: "/**",
  },
  {
    protocol: "http",
    hostname: "127.0.0.1",
    port: "9000",
    pathname: "/**",
  },
  {
    protocol: "http",
    hostname: "localhost",
    port: "4566",
    pathname: "/**",
  },
  {
    protocol: "http",
    hostname: "127.0.0.1",
    port: "4566",
    pathname: "/**",
  },
  {
    protocol: "http",
    hostname: "192.168.1.16",
    port: "9000",
    pathname: "/**",
  },
  {
    protocol: "https",
    hostname: "**.amazonaws.com",
    pathname: "/**",
  },
  {
    protocol: "https",
    hostname: "**.r2.cloudflarestorage.com",
    pathname: "/**",
  },
  {
    protocol: "https",
    hostname: "storage.googleapis.com",
    pathname: "/**",
  },
  {
    protocol: "https",
    hostname: "images.unsplash.com",
    pathname: "/**",
  },
  {
    protocol: "https",
    hostname: "ui-avatars.com",
    pathname: "/**",
  },
];

// Automatically register custom CDN / S3 origins from environment variables without code modifications
for (const envKey of ["NEXT_PUBLIC_MEDIA_BASE_URL", "S3_PUBLIC_URL", "MINIO_PUBLIC_URL"]) {
  const envVal = process.env[envKey]?.trim();
  if (envVal) {
    try {
      const parsed = new URL(envVal.startsWith("http") ? envVal : `https://${envVal}`);
      if (parsed.hostname && !configuredRemotePatterns.some(p => p.hostname === parsed.hostname && p.port === (parsed.port || undefined))) {
        configuredRemotePatterns.push({
          protocol: parsed.protocol.replace(":", "") || "https",
          hostname: parsed.hostname,
          port: parsed.port || undefined,
          pathname: "/**",
        });
      }
    } catch {
      // Ignore unparseable values
    }
  }
}

/** @type {import('next').NextConfig} */
const nextConfig = {
  allowedDevOrigins: ["192.168.1.16"],
  devIndicators: false,
  reactStrictMode: true,
  output: "standalone",
  poweredByHeader: false,
  // Disables lazy compilation in dev mode: keeps compiled pages in memory and disables on-demand proxy compilation
  onDemandEntries: {
    maxInactiveAge: 24 * 60 * 60 * 1000,
    pagesBufferLength: 250,
  },
  webpack(config, { dev }) {
    if (dev && config.experiments) {
      config.experiments.lazyCompilation = false;
    }
    return config;
  },
  headers() {
    return [
      {
        source: "/:path*",
        headers: [
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "X-Frame-Options", value: "SAMEORIGIN" },
          { key: "X-XSS-Protection", value: "1; mode=block" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
          { key: "Permissions-Policy", value: "camera=(), microphone=(), geolocation=()" },
        ],
      },
    ];
  },
  typescript: {
    // Verified by standalone `npx tsc --noEmit` in CI/testing; prevents container build failures
    // caused by alpine node_modules cache artifacts.
    ignoreBuildErrors: true,
  },
  images: {
    remotePatterns: configuredRemotePatterns,
  },
};

export default withNextIntl(nextConfig);
