export default {
  ssr: true,
  build: {
    inlineStylesheets: 'auto',
  },
  image: {
    domains: [],
    service: {
      entrypoint: 'astro/assets/services/sharp',
    },
  },
  experimental: {
    contentLayer: false,
  },
};
