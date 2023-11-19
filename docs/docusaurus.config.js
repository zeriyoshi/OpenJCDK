// @ts-check
// `@type` JSDoc annotations allow editor autocompletion and type checking
// (when paired with `@ts-check`).
// There are various equivalent ways to declare your Docusaurus config.
// See: https://docusaurus.io/docs/api/docusaurus-config

import {themes as prismThemes} from 'prism-react-renderer';

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'OpenJCDK',
  tagline: '邪神ちゃんドロップキック画像botドキュメント',
  favicon: 'img/favicon.ico',

  // Set the production url of your site here
  url: 'https://zeriyoshi.github.io/',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/OpenJCDK',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'zeriyoshi', // Usually your GitHub org/user name.
  projectName: 'OpenJCDK', // Usually your repo name.

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'ja',
    locales: ['ja'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: './sidebars.js',
          editUrl:
            'https://github.com/zeriyoshi/OpenJCDK/tree/main/docs/',
        },
        blog: {
          showReadingTime: true,
          editUrl:
            'https://github.com/zeriyoshi/OpenJCDK/tree/main/docs/'
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      image: 'img/card.jpg',
      navbar: {
        title: 'OpenJCDK',
        logo: {
          alt: 'Logo',
          src: 'img/hebi_aodaisyou.png',
        },
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'usageSidebar',
            position: 'left',
            label: 'ドキュメント',
          },
          { to: '/blog', label: 'ブログ', position: 'left' },
          {
            href: 'https://github.com/zeriyoshi/OpenJCDK',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'More',
            items: [
              {
                label: 'GitHub',
                href: 'https://github.com/zeriyoshi/OpenJCDK',
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} OpenJCDK Project. Built with Docusaurus.`,
      },
      prism: {
        theme: prismThemes.github,
        darkTheme: prismThemes.dracula,
      },
    }),
};

export default config;
