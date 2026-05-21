import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://duyanh-y4n.github.io',
  base: '/dpod-seed',
  integrations: [
    starlight({
      title: 'dpod-seed',
      description: 'CLI for managing reproducible DevPod environments from verified upstream distros.',
      social: [
        { icon: 'github', label: 'GitHub', href: 'https://github.com/iamy4n-dev/dpod-seed' },
      ],
      sidebar: [
        { label: 'Getting Started', link: '/' },
        {
          label: 'Guides',
          items: [
            { label: 'Using dpod-seed', link: '/guides/using-dpod-seed/' },
            { label: 'Writing a distro', link: '/guides/writing-a-distro/' },
            { label: 'Self-hosting', link: '/guides/self-hosting/' },
          ],
        },
        {
          label: 'Reference',
          items: [
            { label: 'CLI commands', link: '/reference/cli-commands/' },
            { label: 'dpod.yaml', link: '/reference/dpod-yaml/' },
            { label: 'distro.yaml', link: '/reference/distro-yaml/' },
            { label: 'dpod.lock', link: '/reference/dpod-lock/' },
          ],
        },
      ],
    }),
  ],
});
