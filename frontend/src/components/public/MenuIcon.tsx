import {
  Cable,
  Circle,
  CircleDot,
  ClipboardList,
  ClipboardPlus,
  Cog,
  Disc3,
  Droplets,
  Factory,
  Gauge,
  Grid2X2,
  Home,
  Info,
  Link as LinkIcon,
  Package,
  Settings,
  Sprout,
  Truck,
  Wheat,
  Wrench,
  type LucideIcon,
} from "lucide-react";

const icons: Record<string, LucideIcon> = {
  home: Home,
  wrench: Wrench,
  cog: Cog,
  truck: Truck,
  droplet: Droplets,
  droplets: Droplets,
  gauge: Gauge,
  disc: Disc3,
  cable: Cable,
  wheat: Wheat,
  sprout: Sprout,
  factory: Factory,
  settings: Settings,
  clipboard: ClipboardList,
  "clipboard-plus": ClipboardPlus,
  grid: Grid2X2,
  package: Package,
  link: LinkIcon,
  info: Info,
  circle: Circle,
  dot: CircleDot,
};

type MenuIconProps = {
  icon?: string;
  className?: string;
  size?: number;
};

export function MenuIcon({ icon, className, size = 18 }: MenuIconProps) {
  const Icon = icon ? icons[icon] || Grid2X2 : Grid2X2;
  return <Icon aria-hidden="true" className={className} size={size} strokeWidth={2} />;
}
