import { defineConfig } from 'vitepress'

export default defineConfig({
    title: 'PIE MDM',
    description: 'Master Data Management(MDM) Platform',
    base: '/piemdm/',
    sitemap: {
        hostname: 'https://pieteams.github.io/piemdm/'
    },
    cleanUrls: true,
    lastUpdated: true,
    ignoreDeadLinks: [
        /^\/sdk\//,
        /localhost/,
    ],
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
            { icon: 'github', link: 'https://github.com/pieteams/piemdm' }
        ],

        footer: {
            message: 'Released under the MIT License.',
            copyright: 'Copyright © 2019-present PieTeams'
        },
    },
    locales: {
        root: {
            label: 'English',
            lang: 'en',
            themeConfig: {
                nav: [
                    {
                        text: 'Getting Started',
                        link: '/getting-started/what-is-piemdm',
                        activeMatch: '/getting-started/',
                    },
                    {
                        text: 'Guide',
                        link: '/guide/user-and-role',
                        activeMatch: '/guide/',
                    },
                    {
                        text: 'Reference',
                        link: '/reference/open-api',
                        activeMatch: '/reference/',
                    },
                ],

                sidebar: {
                    '/getting-started/': [
                        {
                            base: '/getting-started/',
                            text: 'Getting Started',
                            collapsed: false,
                            items: [
                                { text: 'What is PieMdm', link: 'what-is-piemdm' },
                                { text: 'Getting Started', link: 'getting-started' },
                                { text: 'Deploy', link: 'deploy' }
                            ]
                        },
                        {
                            base: '/getting-started/',
                            text: 'Implementation Company List',
                            collapsed: false,
                            items: [
                                { text: 'Chinese Consulting', link: 'china-consulting' },
                                { text: 'International Consulting', link: 'international-consulting' },
                            ]
                        },
                        {
                            base: '/getting-started/',
                            text: 'Master Data System List',
                            collapsed: false,
                            items: [
                                { text: 'Master Data Systems', link: 'master-data-system' },
                            ]
                        },
                    ],
                    '/guide/': [
                        {
                            base: '/guide/',
                            text: 'System Management',
                            collapsed: false,
                            items: [
                                { text: 'User & Role Management', link: 'user-and-role' },
                                { text: 'Permission & Menu Management', link: 'permission-and-menu' },
                            ]
                        },
                        {
                            base: '/guide/',
                            text: 'Data Modeling',
                            collapsed: false,
                            items: [
                                { text: 'Table & Field', link: 'table-and-field' },
                                { text: 'Dictionary', link: 'dictionary' },
                            ]
                        },
                        {
                            base: '/guide/',
                            text: 'Approval Flow',
                            collapsed: false,
                            items: [
                                { text: 'Approval Configuration', link: 'approval-flow' },
                            ]
                        },
                        {
                            base: '/guide/',
                            text: 'Daily Operations',
                            collapsed: false,
                            items: [
                                { text: 'Entity Management', link: 'entity_management' },
                                { text: 'Entity Approval', link: 'entity_approval' },
                            ]
                        },
                    ],
                    '/reference/': [
                        {
                            base: '/reference/',
                            text: 'Reference',
                            items: [
                                { text: 'Open API', link: 'open-api' },
                                { text: 'Call Example', link: 'call-example' },
                                { text: 'SDK', link: 'sdk' },
                            ]
                        }
                    ],
                }
            }
        },
        'zh-CN': {
            label: '简体中文',
            lang: 'zh-CN',
            link: '/zh-CN/',
            themeConfig: {
                nav: [
                    {
                        text: '快速开始',
                        link: '/zh-CN/getting-started/what-is-piemdm',
                        activeMatch: '/zh-CN/getting-started/',
                    },
                    {
                        text: '指南',
                        link: '/zh-CN/guide/user-and-role',
                        activeMatch: '/zh-CN/guide/',
                    },
                    {
                        text: '参考',
                        link: '/zh-CN/reference/open-api',
                        activeMatch: '/zh-CN/reference/',
                    },
                ],
                sidebar: {
                    '/zh-CN/getting-started/': [
                        {
                            base: '/zh-CN/getting-started/',
                            text: '快速开始',
                            collapsed: false,
                            items: [
                                { text: '什么是 PieMdm', link: 'what-is-piemdm' },
                                { text: '开始使用', link: 'getting-started' },
                                { text: '部署', link: 'deploy' }
                            ]
                        },
                        {
                            base: '/zh-CN/getting-started/',
                            text: '实施公司列表',
                            collapsed: false,
                            items: [
                                { text: '中国咨询公司', link: 'china-consulting' },
                                { text: '国际咨询公司', link: 'international-consulting' },
                            ]
                        },
                        {
                            base: '/zh-CN/getting-started/',
                            text: '主数据系统列表',
                            collapsed: false,
                            items: [
                                { text: '主数据系统', link: 'master-data-system' },
                            ]
                        },
                    ],
                    '/zh-CN/guide/': [
                        {
                            base: '/zh-CN/guide/',
                            text: '系统管理',
                            collapsed: false,
                            items: [
                                { text: '用户与角色管理', link: 'user-and-role' },
                                { text: '权限与菜单管理', link: 'permission-and-menu' },
                            ]
                        },
                        {
                            base: '/zh-CN/guide/',
                            text: '数据建模',
                            collapsed: false,
                            items: [
                                { text: '表与字段管理', link: 'table-and-field' },
                                { text: '字典管理', link: 'dictionary' },
                            ]
                        },
                        {
                            base: '/zh-CN/guide/',
                            text: '审批流',
                            collapsed: false,
                            items: [
                                { text: '审批流配置', link: 'approval-flow' },
                            ]
                        },
                        {
                            base: '/zh-CN/guide/',
                            text: '日常操作',
                            collapsed: false,
                            items: [
                                { text: '实体管理', link: 'entity_management' },
                                { text: '审批管理', link: 'entity_approval' },
                            ]
                        },
                    ],
                    '/zh-CN/reference/': [
                        {
                            base: '/zh-CN/reference/',
                            text: '参考',
                            items: [
                                { text: '开放API', link: 'open-api' },
                                { text: '调用示例', link: 'call-example' },
                                { text: 'SDK', link: 'sdk' },
                            ]
                        }
                    ],
                }
            }
        },
        'zh-TW': {
            label: '繁體中文',
            lang: 'zh-TW',
            link: '/zh-TW/',
            themeConfig: {
                nav: [
                    {
                        text: '快速開始',
                        link: '/zh-TW/getting-started/what-is-piemdm',
                        activeMatch: '/zh-TW/getting-started/',
                    },
                    {
                        text: '指南',
                        link: '/zh-TW/guide/user-and-role',
                        activeMatch: '/zh-TW/guide/',
                    },
                    {
                        text: '參考',
                        link: '/zh-TW/reference/open-api',
                        activeMatch: '/zh-TW/reference/',
                    },
                ],
                sidebar: {
                    '/zh-TW/getting-started/': [
                        {
                            base: '/zh-TW/getting-started/',
                            text: '快速開始',
                            collapsed: false,
                            items: [
                                { text: '什麼是 PieMdm', link: 'what-is-piemdm' },
                                { text: '開始使用', link: 'getting-started' },
                                { text: '部署', link: 'deploy' }
                            ]
                        },
                        {
                            base: '/zh-TW/getting-started/',
                            text: '實施公司列表',
                            collapsed: false,
                            items: [
                                { text: '中國諮詢公司', link: 'china-consulting' },
                                { text: '國際諮詢公司', link: 'international-consulting' },
                            ]
                        },
                        {
                            base: '/zh-TW/getting-started/',
                            text: '主數據系統列表',
                            collapsed: false,
                            items: [
                                { text: '主數據系統', link: 'master-data-system' },
                            ]
                        },
                    ],
                    '/zh-TW/guide/': [
                        {
                            base: '/zh-TW/guide/',
                            text: '系統管理',
                            collapsed: false,
                            items: [
                                { text: '用戶與角色管理', link: 'user-and-role' },
                                { text: '權限與菜單管理', link: 'permission-and-menu' },
                            ]
                        },
                        {
                            base: '/zh-TW/guide/',
                            text: '數據建模',
                            collapsed: false,
                            items: [
                                { text: '表與字段管理', link: 'table-and-field' },
                                { text: '字典管理', link: 'dictionary' },
                            ]
                        },
                        {
                            base: '/zh-TW/guide/',
                            text: '審批流',
                            collapsed: false,
                            items: [
                                { text: '審批流配置', link: 'approval-flow' },
                            ]
                        },
                        {
                            base: '/zh-TW/guide/',
                            text: '日常操作',
                            collapsed: false,
                            items: [
                                { text: '實體管理', link: 'entity_management' },
                                { text: '審批管理', link: 'entity_approval' },
                            ]
                        },
                    ],
                    '/zh-TW/reference/': [
                        {
                            base: '/zh-TW/reference/',
                            text: '參考',
                            items: [
                                { text: '開放API', link: 'open-api' },
                                { text: '調用示例', link: 'call-example' },
                                { text: 'SDK', link: 'sdk' },
                            ]
                        }
                    ],
                }
            }
        }
    }
})


