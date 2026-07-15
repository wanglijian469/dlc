import type { Category } from "../types/api";

export function hierarchicalCategoryOptions(categories: Category[]) {
  const sorted = (rows: Category[]) => [...rows].sort((left, right) => (left.sortOrder || 0) - (right.sortOrder || 0) || left.id - right.id);
  const roots = sorted(categories.filter((category) => !category.parentId));
  const children = new Map<number, Category[]>();
  categories.filter((category) => Boolean(category.parentId)).forEach((category) => {
    const parentID = category.parentId || 0;
    children.set(parentID, [...(children.get(parentID) || []), category]);
  });
  const used = new Set<number>();
  const options: Array<{ label: string; value: number }> = [];
  roots.forEach((root) => {
    used.add(root.id);
    options.push({ label: root.name, value: root.id });
    sorted(children.get(root.id) || []).forEach((child) => {
      used.add(child.id);
      options.push({ label: `　└ ${child.name}`, value: child.id });
    });
  });
  sorted(categories.filter((category) => !used.has(category.id))).forEach((category) => options.push({ label: category.name, value: category.id }));
  return options;
}
