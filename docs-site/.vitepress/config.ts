import { defineConfig } from 'vitepress'
import { apiSidebar } from './generated/api-sidebar'
import { guideSidebar } from './generated/guide-sidebar'
import { dartpadPlugin } from './theme/plugins/dartpad'
import { apiLinkerPlugin } from './theme/plugins/api-linker'
import llmstxt, {
  copyOrDownloadAsMarkdownButtons,
} from 'vitepress-plugin-llms'

const docsBase = process.env.DOCS_BASE ?? '/'

const manualGuideItems = [
  {
    text: 'Quick Start Guide',
    link: '/guide/_generated/network_debugger_workspace/quick-start',
  },
  { text: 'Overview', link: '/guide/' },
  { text: 'Firebase Realtime Database', link: '/guide/firebase-database' },
]

const mergedGuideSidebar = {
  '/guide/': [
    {
      text: 'Guide',
      items: manualGuideItems,
    },
    ...(guideSidebar['/guide/'] ?? []),
  ],
}

export default defineConfig({
  base: docsBase,
  title: 'Network Debugger Docs',
  description: 'API и гайды для пакетов из dart_packages',
  ignoreDeadLinks: true,
  lastUpdated: true,
  vite: {
    plugins: [llmstxt()],
  },
  markdown: {
    config: (md) => {
      md.use(dartpadPlugin)
      md.use(apiLinkerPlugin)
      md.use(copyOrDownloadAsMarkdownButtons)
    },
  },
  themeConfig: {
    outline: { level: [2, 4] },
    search: {
      provider: 'local',
    },
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Guide', link: '/guide/' },
      { text: 'API Reference', link: '/api/' },
    ],
    sidebar: {
      ...apiSidebar,
      ...mergedGuideSidebar,
    },
    socialLinks: [
      {
        icon: 'github',
        link: 'https://github.com/cherrypick-agency/flutter_network_debugger',
      },
    ],
  },
})
