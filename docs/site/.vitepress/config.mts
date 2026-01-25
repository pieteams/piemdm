import { defineConfig } from 'vitepress'

export default defineConfig({
    title: 'PIE MDM',
    description: 'Master Data Management(MDM) Platform',
    base: '/piemdm/',
    sitemap: {
        hostname: 'https://pieworks.github.io/piemdm/'
    },
    head: [
        ['script', { async: '', src: 'https://www.googletagmanager.com/gtag/js?id=G-LBRJN3MR40' }],
        ['script', {}, `
      window.dataLayer = window.dataLayer || [];
      function gtag(){dataLayer.push(arguments);}
      gtag('js', new Date());
      gtag('config', 'G-LBRJN3MR40');
    `]
    ],
    themeConfig: {
        socialLinks: [
            { icon: 'github', link: 'https://github.com/pieworks/piemdm' }
        ],

        footer: {
            message: 'Released under the MIT License.',
            copyright: 'Copyright © 2019-present PieWorks'
        },
    },
    locales: {
        root: {
            label: 'English',
            lang: 'en',
            themeConfig: {
                nav: [
                    { text: 'Guide', link: '/guide/getting-started' },
                    { text: 'Reference', link: '/reference/api' }
                ],
                sidebar: [
                    {
                        text: 'Guide',
                        items: [
                            { text: 'Getting Started', link: '/guide/getting-started' }
                        ]
                    },
                    {
                        text: 'Reference',
                        items: [
                            { text: 'API Reference', link: '/reference/api' }
                        ]
                    }
                ]
            }
        },
        'zh-CN': {
            label: '简体中文',
            lang: 'zh-CN',
            link: '/zh-CN/',
            themeConfig: {
                nav: [
                    { text: '指南', link: '/zh-CN/guide/getting-started' },
                    { text: '参考', link: '/zh-CN/reference/api' }
                ],
                sidebar: [
                    {
                        text: '指南',
                        items: [
                            { text: '快速开始', link: '/zh-CN/guide/getting-started' }
                        ]
                    },
                    {
                        text: '参考',
                        items: [
                            { text: 'API 参考', link: '/zh-CN/reference/api' }
                        ]
                    }
                ]
            }
        },
        'zh-TW': {
            label: '繁體中文',
            lang: 'zh-TW',
            link: '/zh-TW/',
            themeConfig: {
                nav: [
                    { text: '指南', link: '/zh-TW/guide/getting-started' },
                    { text: '參考', link: '/zh-TW/reference/api' }
                ],
                sidebar: [
                    {
                        text: '指南',
                        items: [
                            { text: '快速開始', link: '/zh-TW/guide/getting-started' }
                        ]
                    },
                    {
                        text: '參考',
                        items: [
                            { text: 'API 參考', link: '/zh-TW/reference/api' }
                        ]
                    }
                ]
            }
        }
    }
})