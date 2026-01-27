import {
    PhotoIcon,
    CloudArrowUpIcon,
    Cog6ToothIcon,
    MapIcon
} from "@heroicons/vue/24/outline";
import type { FunctionalComponent, HTMLAttributes, VNodeProps } from 'vue'

export interface NavChildLink {
    path: string
    label: string
}

export interface BaseNavLink {
    path: string
    label: string
    defaultPath?: string
    alwaysInMore?: boolean
    icon?: FunctionalComponent<HTMLAttributes & VNodeProps>
}

export interface NavLinkWithChildren extends BaseNavLink {
    children: NavChildLink[]
}

export type NavLink = BaseNavLink | NavLinkWithChildren

export const useNavLinks = (): NavLink[] => {
    return [
        {
            path: '/gallery',
            label: 'nav.gallery',
            icon: PhotoIcon
        },
        {
            path: '/upload',
            label: 'nav.upload',
            icon: CloudArrowUpIcon
        },
        {
            path: '/routes',
            label: 'nav.routes',
            icon: MapIcon
        },
        {
            path: '/settings',
            label: 'nav.settings',
            icon: Cog6ToothIcon
        }
    ]
}
