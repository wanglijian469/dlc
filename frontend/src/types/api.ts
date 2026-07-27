export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface Menu {
  id: number;
  name: string;
  parentId?: number;
	categoryId?: number;
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

export interface VendorMedia {
  id?: number;
  vendorId?: number;
  kind: "factory" | "equipment" | "certificate";
  url: string;
  caption?: string;
  sortOrder?: number;
	assetId?: number;
}

export interface Vendor {
  id: number;
  slug?: string;
  name: string;
  shortName?: string;
  logo?: string;
	logoAssetId?: number;
  coverImage?: string;
	coverAssetId?: number;
  province?: string;
  city?: string;
  county?: string;
  address?: string;
  mainProducts?: string;
  serviceModels?: string;
  serviceAdvantages?: string;
  description?: string;
  seoTitle?: string;
  seoDescription?: string;
  seoTitleManual?: boolean;
  seoDescriptionManual?: boolean;
  establishedYear?: string;
  factoryArea?: string;
  employeeCount?: string;
  annualCapacity?: string;
  equipment?: string;
  certifications?: string;
  afterSalesService?: string;
  reviewStatus?: "pending" | "verified" | "rejected" | string;
  providesProcessing?: boolean;
  processingServices?: string;
  processingMaterials?: string;
  processingEquipment?: string;
  processingCapacity?: string;
  processingRegions?: string;
  processingNotes?: string;
  websiteUrl?: string;
  phone?: string;
  wechat?: string;
  contactName?: string;
  isRecommended?: boolean;
  isVerified?: boolean;
  isVisible?: boolean;
	dataOrigin?: "vendor_submission" | "verified_source" | "admin" | "demo";
	publicationStatus?: "draft" | "in_review" | "scheduled" | "published" | "rejected" | "archived" | "hidden";
	contentVersion?: number;
	publishedAt?: string;
  sortOrder?: number;
  tags?: Tag[];
  media?: VendorMedia[];
  tagIds?: number[];
}

export interface VendorOption {
  id: number;
  name: string;
  shortName?: string;
  province?: string;
  city?: string;
  mainProducts?: string;
  publicationStatus?: Vendor["publicationStatus"];
  isVisible?: boolean;
}

export interface Category {
  id: number;
  slug?: string;
  name: string;
  parentId?: number;
  icon?: string;
  seoTitle?: string;
  seoDescription?: string;
  sortOrder?: number;
  isEnabled?: boolean;
  publicationStatus?: "draft" | "in_review" | "scheduled" | "published" | "rejected" | "archived";
  contentVersion?: number;
  publishedAt?: string;
}

export interface ProductSpec {
  name: string;
  value: string;
}

export interface Product {
  id: number;
  slug?: string;
  name: string;
  image?: string;
  categoryId?: number;
  vendorId?: number;
  compatibleModels?: string;
  description?: string;
  detailContent?: string;
  seoTitle?: string;
  seoDescription?: string;
  galleryRaw?: string;
  gallery?: string[];
  specsRaw?: string;
  specs?: ProductSpec[];
  priceNote?: string;
  inquiryText?: string;
  inquiryPath?: string;
  isHot?: boolean;
  isRecommended?: boolean;
  sortOrder?: number;
  status?: number;
  publicationStatus?: "draft" | "in_review" | "scheduled" | "published" | "rejected" | "archived" | "hidden";
  contentVersion?: number;
  publishedAt?: string;
  supplierCount?: number;
  supplierRegions?: string[];
  supplier?: ProductSupplier;
  associationCount?: number;
  associatedVendors?: VendorOption[];
  category?: Category;
  vendor?: Vendor;
}

export interface ProductSupplier {
  id: number;
  productId: number;
  vendorId: number;
  vendorProductName?: string;
  vendorModel?: string;
  image?: string;
  galleryRaw?: string;
  gallery?: string[];
  compatibleModels?: string;
  description?: string;
  priceNote?: string;
  inquiryText?: string;
  inquiryPath?: string;
  sourceType?: string;
  status: "pending" | "approved" | "rejected" | "disabled";
  reviewNote?: string;
  contentVersion?: number;
  product?: Product;
  vendor?: Vendor;
}

export interface ProductSubmission {
  id: number;
  vendorId: number;
  productId?: number;
  supplierId?: number;
  submissionType: "new_product" | "link_product" | "link_supplier" | "update_supplier" | "update_offer";
  baseVersion?: number;
  status: "pending" | "approved" | "rejected" | "superseded";
  reviewNote?: string;
  submittedBy?: string;
  reviewedBy?: string;
  productDraft?: Partial<Product>;
  supplierDraft?: Partial<ProductSupplier>;
  product?: Product;
  supplier?: ProductSupplier;
  vendor?: Vendor;
  createdAt: string;
  updatedAt: string;
}

export interface VendorProductRecord {
  product: Product;
  supplier: ProductSupplier;
  latestSubmission?: ProductSubmission;
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

export interface SiteMeta {
  siteName: string;
  brandMark: string;
  brandLogo?: string;
  submitVendorText: string;
  adminLoginText: string;
  mobileBrandName: string;
  mobileBrandMark: string;
  copyrightOwner?: string;
  copyrightYear?: string;
  filingNumber?: string;
  siteUrl?: string;
  defaultSeoTitle?: string;
  defaultSeoDescription?: string;
  baiduVerification?: string;
  googleVerification?: string;
}

export interface ThemeConfig { primaryColor: string; accentColor: string; }
export interface LayoutConfig {
  siteMeta: SiteMeta;
  theme: ThemeConfig;
  topMenus: Menu[];
  sidebarMenus: Menu[];
  auxiliaryMenus: Menu[];
  mobileMenus: Menu[];
  mobileBottomMenus: Menu[];
  version: string;
}

export interface HomeSections {
  recommendedTitle?: string;
  recommendedLink?: string;
  moreTitle?: string;
  moreLink?: string;
  recommendedLimit?: number;
  moreLimit?: number;
  showRecommended?: boolean;
  showMore?: boolean;
}

export type HomeModuleType = "recommendedVendors" | "processingServices" | "moreVendors";
export interface HomeModule {
  type: HomeModuleType;
  title: string;
  subtitle?: string;
  visible: boolean;
  limit?: number;
  path?: string;
  image?: string;
  sortOrder: number;
}

// Kept only for source compatibility with retired components; the public home
// endpoint no longer emits or renders this configuration.
export interface JoinConfig {
  text: string;
  buttonText: string;
  path: string;
}

export interface ContentPageRecord {
  id: number;
  slug: string;
  title: string;
  summary?: string;
  content?: string;
  blocksRaw?: string;
  blocks?: ContentBlock[];
  seoKeywords?: string;
  seoTitle?: string;
  seoDescription?: string;
  pageType?: "page" | "article";
  coverImage?: string;
  authorName?: string;
  publishedAt?: string;
  relatedCategoryId?: number;
  relatedProductId?: number;
  relatedVendorId?: number;
  isEnabled?: boolean;
  sortOrder?: number;
}

export type ContentBlockType = "hero" | "text" | "steps" | "cta" | "contact" | "faq";

export interface ContentBlock {
  type: ContentBlockType;
  title?: string;
  text?: string;
  items?: string[];
  buttonText?: string;
  buttonPath?: string;
  phone?: string;
  wechat?: string;
}

export interface FriendLink {
  id: number;
  name: string;
  url: string;
  logo?: string;
  sortOrder?: number;
  isEnabled?: boolean;
}

export interface StatItem {
  label: string;
  value: string;
}

export interface HomePayload {
  siteMeta?: SiteMeta;
  topMenus: Menu[];
  sidebarMenus: Menu[];
  auxiliaryMenus: Menu[];
  mobileMenus: Menu[];
  mobileBottomMenus?: Menu[];
  banner: Banner;
  recommendedVendors: Vendor[];
  moreVendors: Vendor[];
  stats: StatItem[];
  modules?: HomeModule[];
  processingVendors?: Vendor[];
  /** @deprecated Retired homepage configuration, not rendered. */
  safeguards?: string[];
  /** @deprecated Retired homepage configuration, not rendered. */
  join?: JoinConfig;
  /** @deprecated Retired homepage payload, not rendered. */
  popularCategories?: Category[];
  /** @deprecated Retired homepage payload, not rendered. */
  featuredProducts?: Product[];
}

export interface PageResult<T> {
  items: T[];
  page: number;
  pageSize: number;
  total: number;
}

export interface FilterOptions {
  provinces: string[];
  categories: Category[];
  serviceTags: Tag[];
}

export interface SearchPayload {
  vendors: PageResult<Vendor>;
  products: PageResult<Product>;
  categories: PageResult<Category>;
}
