import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'
import { apiSidebar } from './generated/api-sidebar'
import { guideSidebar } from './generated/guide-sidebar'
import { dartpadPlugin } from './theme/plugins/dartpad'
import { apiLinkerPlugin } from './theme/plugins/api-linker'
import llmstxt from 'vitepress-plugin-llms'

const docsBase = process.env.DOCS_BASE ?? '/'

// Organize sidebar into logical sections
const organizeSidebar = () => {
  const generated = guideSidebar['/guide/']?.[0]?.items || []

  // Extract packages
  const packages = generated.filter((item: any) =>
    item.link?.includes('/packages/')
  ).map((item: any) => ({
    text: item.text,
    link: item.link
  }))

  return [
    {
      text: 'Getting Started',
      items: [
        { text: 'Overview', link: '/guide/' },
        { text: 'Quick Start', link: '/guide/network_debugger_workspace/quick-start' },
        { text: 'Desktop Setup', link: '/guide/network_debugger_workspace/desktop-setup' },
      ]
    },
    {
      text: 'Configuration',
      items: [
        { text: 'Configuration Guide', link: '/guide/network_debugger_workspace/configuration' },
        { text: 'Proxy Modes', link: '/guide/network_debugger_workspace/proxy-modes' },
        { text: 'Platform Support', link: '/guide/network_debugger_workspace/platform-support' },
      ]
    },
    {
      text: 'Packages',
      collapsed: true,
      items: packages
    },
    {
      text: 'Specialized Debuggers',
      items: [
        { text: 'Firebase Realtime Database', link: '/guide/firebase-database' },
      ]
    },
    {
      text: 'Reference',
      items: [
        { text: 'API Reference', link: '/guide/network_debugger_workspace/api-reference' },
        { text: 'Troubleshooting', link: '/guide/network_debugger_workspace/troubleshooting' },
      ]
    },
  ]
}

const mergedGuideSidebar = {
  '/guide/': organizeSidebar(),
}

export default withMermaid(defineConfig({
  base: docsBase,
  title: 'Network Debugger Docs',
  description: 'API и гайды для пакетов из dart_packages',
  ignoreDeadLinks: true,
  lastUpdated: true,
  vite: {
    plugins: [llmstxt()],
    optimizeDeps: {
      include: ['mermaid'],
    },
    ssr: {
      noExternal: ['mermaid'],
    },
  },
  markdown: {
    config: (md) => {
      md.use(dartpadPlugin)
      md.use(apiLinkerPlugin)
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
}))
