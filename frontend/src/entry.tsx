import Astro from 'astro:server';

export async function render(props: { url: URL; site?: URL; prerender: boolean }) {
  return Astro.render(props);
}
