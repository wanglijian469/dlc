export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface Menu {
  id: number;
  name: string;
  parentId?: number;
  icon?: string;
  menuType?: string;
  path?: string;
  sortOrder?: number;
  isEnabled?: boolean;
  isTop?: boolean;
  isDefaultOpen?: boolean;
  children?: Menu[];
}

export interface Tag {
  id: number;
  name: string;
  tagType?: string;
  color?: string;
  sortOrder?: number;
}

export interface Vendor {
  id: number;
  name: string;
  shortName?: string;
  logo?: string;
  coverImage?: string;
  province?: string;
  city?: string;
  county?: string;
  address?: string;
  mainProducts?: string;
  serviceModels?: string;
  serviceAdvantages?: string;
  description?: string;
  websiteUrl?: string;
  phone?: string;
  wechat?: string;
  contactName?: string;
  isRecommended?: boolean;
  isVerified?: boolean;
  isVisible?: boolean;
  sortOrder?: number;
  tags?: Tag[];
}

export interface Category {
  id: number;
  name: string;
  parentId?: number;
  icon?: string;
  sortOrder?: number;
  isEnabled?: boolean;
}

export interface Product {
  id: number;
  name: string;
  image?: string;
  categoryId?: number;
  vendorId?: number;
  compatibleModels?: string;
  description?: string;
  isHot?: boolean;
  isRecommended?: boolean;
  sortOrder?: number;
  status?: number;
  category?: Category;
  vendor?: Vendor;
}

export interface Banner {
  id?: number;
  title: string;
  subtitle?: string;
  backgroundImage?: string;
  searchPlaceholder?: string;
  hotKeywords?: string[];
  hotKeywordsRaw?: string;
  isEnabled?: boolean;
  sortOrder?: number;
}

export interface SiteConfig {
  id: number;
  configKey: string;
  configValue: string;
  description?: string;
}

export interface StatItem {
  label: string;
  value: string;
}

export interface JoinConfig {
  text: string;
  buttonText: string;
  path: string;
}

export interface HomePayload {
  topMenus: Menu[];
  sidebarMenus: Menu[];
  auxiliaryMenus: Menu[];
  mobileMenus: Menu[];
  banner: Banner;
  recommendedVendors: Vendor[];
  moreVendors: Vendor[];
  stats: StatItem[];
  safeguards: string[];
  join: JoinConfig;
}
